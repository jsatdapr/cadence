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
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Fuzzing tools often find new samples by simply appending
// bytes to the old samples.  Assert here that adding new bytes
// doesn't change the output, until all the old output is used up.
func TestThatFuzzbitsCanBeAppendedTo(t *testing.T) {
	seed := 123456 + time.Now().Unix()
	rnd := rand.New(rand.NewSource(seed))
	oldSample := make([]byte, 1+rnd.Intn(30))
	appendage := make([]byte, 1+rnd.Intn(10))
	rnd.Read(oldSample)
	rnd.Read(appendage)
	fbExpected := NewFuzzbits(oldSample)
	fbActual := NewFuzzbits(append(oldSample, appendage...))
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
