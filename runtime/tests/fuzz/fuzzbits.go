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
	"math/bits"
)

type Fuzzbits interface {
	BitsLeft() int
	Intn(n int) int
}

type defaultFuzzbits struct {
	data   []byte
	offset int
}

func NewFuzzbits(chunkSize int, data []byte) Fuzzbits {
	if chunkSize == 8 {
		return &defaultFuzzbits{data: data}
	} else {
		return NewChunkedFuzzbits(chunkSize, data)
	}
}

func (x *defaultFuzzbits) BitsLeft() int {
	return (len(x.data) - x.offset) * 8
}

func (x *defaultFuzzbits) Intn(n int) int {
	if n <= 0 {
		panic(fmt.Errorf("n %d, nl %d", n, x.BitsLeft()))
	}
	if n == 1 {
		return 0
	}
	if x.BitsLeft() <= 0 {
		x.offset++
		return x.offset % n
	}
	result := int(x.data[x.offset])
	x.offset++
	for p := n; p > 256; p /= 256 {
		result <<= 8
		result += x.Intn(256)
	}
	return result % n
}

//////////////////////////////////////////////////////////////////////////////////////////////
// based on https://fgiesen.wordpress.com/2018/09/27/reading-bits-in-far-too-many-ways-part-3/
//
type chunkedFuzzbits struct {
	data      []byte
	offset    int
	bits      uint64
	chunkSize int
}

func NewChunkedFuzzbits(chunkSize int, data []byte) Fuzzbits {
	return &chunkedFuzzbits{data: data, bits: 1 << 63, chunkSize: chunkSize}
}

func (r *chunkedFuzzbits) BitsLeft() int {
	return (len(r.data)-r.offset)*8 - bits.LeadingZeros64(r.bits)
}

func (r *chunkedFuzzbits) Intn(n int) int {
	minimumBits := 32 - bits.LeadingZeros32(uint32(n-1))
	chunkedBits := ((minimumBits + r.chunkSize - 1) / r.chunkSize) * r.chunkSize
	if n <= 0 {
		panic(fmt.Errorf("n %d, nl %d", n, r.BitsLeft()))
	}
	if n == 1 {
		return 0
	}
	if r.BitsLeft() <= 0 {
		r.offset++
		return r.offset % n
	}
	if chunkedBits > r.BitsLeft() {
		chunkedBits = r.BitsLeft()
	}
	return int(r.GetBits(chunkedBits)) % n
}

func (r *chunkedFuzzbits) GetBits(count int) uint64 {
	if count < 1 || count > 56 {
		panic(count)
	}
	bits_consumed := bits.LeadingZeros64(r.bits) //"Count how many bits we consumed
	r.offset += bits_consumed >> 3               // Advance the pointer
	r.bits = r.current() | (1 << 63)             // Refill and put the marker in the MSB
	r.bits >>= bits_consumed & 7                 // Consume the bits in this byte that we've already used.
	x := r.bits & ((1 << count) - 1)             // Just need to mask the low bits." - fgiesen
	r.bits >>= count
	return x
}

func (r *chunkedFuzzbits) current() uint64 {
	result := uint64(0)
	end := r.offset + 8
	if end >= len(r.data) {
		end = len(r.data) - 1
	}
	for i := end; i >= r.offset; i-- {
		result <<= 8
		result |= uint64(r.data[i])
	}
	return result
}
