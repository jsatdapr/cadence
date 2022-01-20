//go:build fuzzbuzz
// +build fuzzbuzz

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

import "git.fuzzbuzz.io/fuzz"

func FuzzRandomBytes(f *fuzz.F)             { runByteSample(f.Bytes("bs").Get()) }
func FuzzRandomStrings(f *fuzz.F)           { runStringSample(f.String("ss").Get()) }
func FuzzRandomTokenStream(f *fuzz.F)       { runRandomTokenStreamSample(8, f.Bytes("bs").Get()) }
func FuzzSimpleRandomTokenStream(f *fuzz.F) { runSimpleRandomTokenStreamSample(8, f.Bytes("bs").Get()) }
