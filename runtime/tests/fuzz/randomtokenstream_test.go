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
	"fmt"
	"hash/fnv"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/cadence/runtime/parser2/lexer"
)

type BigSetOfHashes struct {
	bitSetHI big.Int
	bitSetLO big.Int
}

func (x *BigSetOfHashes) Add(s string) bool {
	h := fnv.New64a()
	h.Write([]byte(s))
	return x.AddHash(h.Sum64())
}

func (x *BigSetOfHashes) AddHash(hash uint64) bool {
	hashHI := int((hash >> 32) & 0x7fffffff)
	hashLO := int(hash & 0x7fffffff)
	if x.bitSetLO.Bit(hashLO) == 0 || x.bitSetHI.Bit(hashHI) == 0 {
		x.bitSetLO.SetBit(&x.bitSetLO, hashLO, 1)
		x.bitSetHI.SetBit(&x.bitSetHI, hashHI, 1)
		return true
	}
	return false
}

func mkRandomTokenStream(data []byte) lexer.TokenStream {
	return &RandomTokenStream{Fuzzbits: NewFuzzbits(data)}
}

func mkSimpleRandomTokenStream(data []byte) lexer.TokenStream {
	return &SimpleRandomTokenStream{Fuzzbits: NewFuzzbits(data)}
}

func TestRandomTokenStreams(t *testing.T) {
	total := 1000000
	pct := func(x int) float64 { return float64(x) * 100.0 / float64(total) }
	fmt.Printf("%7s,%7s,%7s,%7s ... %7s,%7s,%7s,%7s,%7s\n",
		"fail%", "empty%", "ok%", "dupe%", "total#",
		"fail#", "empty#", "ok#", "dupe#")

	for name, fun := range map[string]func([]byte) lexer.TokenStream{
		"RandomTokenStream":       mkRandomTokenStream,
		"SimpleRandomTokenStream": mkSimpleRandomTokenStream,
	} {
		fail, empty, ok, dupe := testTokenStream(total, fun)
		fmt.Printf("%6.2f%%,%6.2f%%,%6.2f%%,%6.2f%% ... %7d,%7d,%7d,%7d,%7d ... %s\n",
			pct(fail), pct(empty), pct(ok), 100.*pct(dupe)/(pct(ok+dupe)), total, fail, empty, ok, dupe, name)
	}
}

func testTokenStream(total int, mkstream func(data []byte) lexer.TokenStream) (fail, empty, ok, dupe int) {
	dupes := BigSetOfHashes{}
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < total; i++ {
		data := make([]byte, 3+rnd.Intn(5))
		rnd.Read(data)
		rts := &lexer.CodeGatheringTokenStream{
			Delegate: &lexer.AutoSpacingTokenStream{
				Delegate: mkstream(data)}}
		program, _ := parser2.ParseProgramFromTokenStream(rts)
		if program == nil {
			fail++
		} else if len(program.Declarations()) == 0 {
			empty++
		} else if !dupes.Add(rts.Input()) {
			dupe++
		} else {
			ok++
		}
	}

	return
}
