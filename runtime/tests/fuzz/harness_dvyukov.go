//go:build !gofuzzbeta && !fuzzbuzz
// +build !gofuzzbeta,!fuzzbuzz

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

func FuzzRandomBytes(data []byte) int             { return runByteSample(data) }
func FuzzRandomTokenStream(data []byte) int       { return runRandomTokenStreamSample(8, data) }
func FuzzSimpleRandomTokenStream(data []byte) int { return runSimpleRandomTokenStreamSample(8, data) }
