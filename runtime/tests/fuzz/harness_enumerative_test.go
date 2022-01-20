/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2022 Dapper Labs, Inc.
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
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

func testEnumeratively(runner func(int, []byte) int) {
	fuzzUntil := time.Now().Unix() + 10
	if ft, err := strconv.Atoi(os.Getenv("FUZZTIME")); err == nil && ft != 0 {
		fuzzUntil = time.Now().Unix() + int64(ft)
	}

	data := make([]byte, 8)
	start := uint64(rand.New(rand.NewSource(time.Now().UnixNano())).Int63())
	fmt.Printf("testEnumeratively: start = %d\n", start)
	for i := uint64(0); ; i++ {
		binary.LittleEndian.PutUint64(data, start+i)

		randomLength := 3 + start%5
		runner(0, data[0:randomLength])

		// every once in a while ...
		if i != 0 && 0 == i&((1<<14)-1) {
			if time.Now().Unix() > fuzzUntil {
				break // ... check time, exit if done.
			}
			// and, if not already showing stats ...
			if !FUZZSTATS { // ... show some progress
				fmt.Print(".")
				if 0 == i&((1<<20)-1) {
					fmt.Println()
				}
			}
		}
	}
	fmt.Println()
}

func TestFuzzRandomTokenStreamEnumeratively(t *testing.T) {
	testEnumeratively(runRandomTokenStreamSample)
}

func TestFuzzSimpleRandomTokenStreamEnumeratively(t *testing.T) {
	testEnumeratively(runSimpleRandomTokenStreamSample)
}
