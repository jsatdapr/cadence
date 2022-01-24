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
)

type Fuzzbits interface {
	BitsLeft() int
	Intn(n int) int
}

type defaultFuzzbits struct {
	data   []byte
	offset int
}

func NewFuzzbits(data []byte) Fuzzbits {
	return &defaultFuzzbits{data: data}
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
