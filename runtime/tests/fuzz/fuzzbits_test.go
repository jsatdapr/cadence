/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fuzz

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThatBignumFuzzbitsWorkAsExpected(t *testing.T) {
	millisPerYear := 365 * 24 * 60 * 60 * 1000

	yearday, hour, minute, second, milliseconds := 0, 17, 23, 59, 123
	pointInTime := 0 +
		yearday*24*60*60*1000 +
		hour /**/ *60*60*1000 +
		minute /* */ *60*1000 +
		second /*    */ *1000 + milliseconds

	mid := big.NewRat(int64(pointInTime), 1)
	max := big.NewRat(int64(millisPerYear), 1)
	fuzzbits := bignumFuzzbits{mid: mid, max: max}
	assert.NotZero(t, fuzzbits.BitsLeft())
	assert.Equal(t, 0, fuzzbits.Intn(365))
	assert.Equal(t, 17, fuzzbits.Intn(24))
	assert.Equal(t, 23, fuzzbits.Intn(60))
	assert.Equal(t, 59, fuzzbits.Intn(60))
	assert.Equal(t, 123, fuzzbits.Intn(1000))
	assert.Zero(t, fuzzbits.BitsLeft())
	// when fuzzbits run out, you could just return infinite 0 and call that sensible.
	// unfortunately, decision-based generators are usually recursive, and infinite 0
	// ends in stack overflow.  instead return infinite incrememnt, and hope that helps.
	for j := 0; j < 12345; j++ {
		assert.Equal(t, j, fuzzbits.Intn(j+1))
	}
}

// Fuzzing tools often find new samples by simply appending
// bytes to the old samples.  Assert here that adding new bytes
// doesn't change the output, until all the old output is used up.
func TestThatFuzzbitsCanBeAppendedTo(t *testing.T) {
	for chunkSize := 0; chunkSize <= 17; chunkSize++ {
		testThatFuzzbitsCanBeAppendedTo(t, chunkSize, rand.Int63())
	}
}

func testThatFuzzbitsCanBeAppendedTo(t *testing.T, chunkSize int, seed int64) {
	rnd := rand.New(rand.NewSource(seed))
	oldSample := make([]byte, 1+rnd.Intn(30))
	appendage := make([]byte, 1+rnd.Intn(10))
	rnd.Read(oldSample)
	rnd.Read(appendage)
	fbExpected := NewFuzzbits(chunkSize, oldSample)
	fbActual := NewFuzzbits(chunkSize, append(oldSample, appendage...))
	for {
		modulus := 1 + rnd.Intn(256)
		expected := fbExpected.Intn(modulus)
		if fbExpected.BitsLeft() <= 0 {
			break
		}
		actual := fbActual.Intn(modulus)
		if !assert.Equal(t, expected, actual, "seed %d\n", seed) {
			t.FailNow()
		}
	}
}

// consuming fuzzbits in chunks (even 1 bit) wastes bits and gives duplicate outputs.
// (and that modular bignum fuzzbits waste 0 bits, giving zero duplicate outputs)
// i.e. waste 0bits: 0% dupe, 0.5b 25% dupes, 1b 50% dupes, 2b 75% dupes ... etc.
func TestThatChunkedFuzzbitsProduceMostlyDuplicates(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Parallel()

	rndSampling := func(data []byte, _ int) { rand.Read(data) }
	enumerative := func(data []byte, i int) { binary.LittleEndian.PutUint32(data, uint32(i)) }

	for name, moduli := range map[string][]int{
		"hourofday": {24},
		"minofhour": {60},
		"dayofyear": {365},
		"minofday1": {24, 60},
		"minofday2": {60, 24},
		"pik2cards": {52, 51},
		"furlongft": {10, 22, 3},
	} {
		for chunkSize := 0; chunkSize <= 8; chunkSize++ {
			accuracy := 0.000001 // enumerating every possible input one time, gives precise answer
			expected, actual := calcDupePct(chunkSize, moduli, enumerative, func(float64) int { return 1 })
			assert.InDelta(t, expected, actual, accuracy, "enum %s (%d)", name, chunkSize)

			accuracy = 1.0 // random sampling; multiple rounds to prove prediction to 1% accuracy
			numRounds := func(waste float64) int { return int((100. / accuracy) / (waste * waste / math.E)) }
			expected, actual = calcDupePct(chunkSize, moduli, rndSampling, numRounds)
			assert.InDelta(t, expected, actual, accuracy, "rand %s (%d) %d", name, chunkSize, numRounds)

			if chunkSize == 0 { // modular bignums waste zero bits...
				assert.Zero(t, expected) // ... so expect zero dupes.
			}
		}
	}
}

func calcDupePct(chunkSize int, moduli []int, getbits func([]byte, int), numRoundsF func(float64) int) (float64, float64) {
	totalWastedBits := 0.
	totalUsedBits := 0
	differentPossibleOutputs := 1

	fb := NewFuzzbits(chunkSize, make([]byte, 100))
	for _, N := range moduli {
		differentPossibleOutputs *= N

		if chunkSize == 0 {
			continue
		}

		numFractionalBitsNeededToSelectN := math.Log2(float64(N))
		numIntegralBitsNeededToSelectN := int(math.Ceil(numFractionalBitsNeededToSelectN))
		numChunkedBitsNeededToSelectN := chunkSize * ((numIntegralBitsNeededToSelectN + chunkSize - 1) / chunkSize)
		wastedBitsOnSelectingN := float64(numChunkedBitsNeededToSelectN) - numFractionalBitsNeededToSelectN

		had, _, have := fb.BitsLeft(), fb.Intn(N), fb.BitsLeft()
		actualUsedBits := had - have
		if actualUsedBits != numChunkedBitsNeededToSelectN {
			panic(fmt.Errorf("expected %d, actual %d\n", numChunkedBitsNeededToSelectN, actualUsedBits))
		}

		totalUsedBits += actualUsedBits
		totalWastedBits += wastedBitsOnSelectingN
	}

	if totalUsedBits > 26 {
		panic(fmt.Errorf("too many bits for this test harness: %d\n", totalUsedBits))
	}

	// lets try every different possible input ...
	differentPossibleInputs := 1 << totalUsedBits

	// ... and see how many duplicates that produces.
	// we predict XX%, where XX=100*(1.0-0.5^wastedBits)
	// i.e. 1 wasted bit=>50% dupes, 2 wasted bits=>75% dupes, ...
	expectedDupePct := 100.0 * (1.0 - math.Pow(1/2., totalWastedBits))

	// so let's keep a count for each possible output while we ...
	counts := make([]int, differentPossibleOutputs)
	input := make([]byte, 4)                 // ... try every possible input ...
	numRounds := numRoundsF(totalWastedBits) // ... "numRounds" times.
	numSamples := differentPossibleInputs * numRounds
	for i := 0; i < numSamples; i++ {
		getbits(input, i)
		fb := NewFuzzbits(chunkSize, input)
		choice := 0
		for _, N := range moduli {
			choice = choice*N + fb.Intn(N)
		}
		counts[choice]++
	}

	actualDupes := 0
	expectedCount := numRounds
	for _, count := range counts {
		if count > expectedCount {
			actualDupes += count - expectedCount
		}
	}

	actualDupePct := float64(100*actualDupes) / float64(numSamples)

	return expectedDupePct, actualDupePct
}
