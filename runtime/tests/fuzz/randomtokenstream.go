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
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/cadence/runtime/parser2/lexer"
)

type RandomTokenStream struct {
	Data        []byte
	dataIndex   int
	accumulator int
	sentKeyword bool
}

func (ts *RandomTokenStream) Input() string {
	panic("unimplemented")
}

func (ts *RandomTokenStream) intn(n int) int {
	if n <= 0 || n >= 256 {
		panic(errors.NewUnreachableError())
	}
	b := ts.Data[ts.dataIndex%len(ts.Data)]
	ts.dataIndex++
	if ts.dataIndex >= len(ts.Data) {
		ts.accumulator++
	}
	return (int(b) + ts.accumulator) % n
}

func (ts *RandomTokenStream) Next() lexer.Token {
	if len(ts.Data) == 0 { // sometimes the fuzzing harness gives us zero bytes
		return lexer.Token{Type: lexer.TokenEOF}
	}

	if ts.accumulator > 0 { // reached end of random fuzz bits?
		if ts.intn(3) == 1 { // prefer EOF now
			return lexer.Token{Type: lexer.TokenEOF}
		}
	}

	if !ts.sentKeyword {
		ts.sentKeyword = true
		firstWhat := ts.intn(4)
		t := lexer.Token{Type: lexer.TokenIdentifier}
		if firstWhat <= 1 { // 50% of the time, keyword
			t.Value = parser2.Keywords[ts.intn(len(parser2.Keywords))]
		} else if firstWhat == 2 { // 25%, start with some other identifier
			ids := []string{"a", "b", "c", "d"}
			t.Value = ids[ts.intn(len(ids))]
		} else if firstWhat == 3 { // 25%, start with a pragma
			t.Type = lexer.TokenPragma
		}
		return t
	}

	min := int(lexer.TokenBinaryIntegerLiteral)
	max := int(lexer.TokenPragma)
	ty := lexer.TokenType(min + ts.intn(max-min+1))

	t := lexer.Token{Type: ty}
	if ty == lexer.TokenIdentifier {
		t.Value = []string{"a", "b", "c", "d"}[ts.intn(4)]
	} else if ty == lexer.TokenString {
		t.Value = "\"string\""
	} else if ty == lexer.TokenFixedPointNumberLiteral {
		t.Value = []string{"0.1", "1.0"}[ts.intn(2)]
	} else if ty == lexer.TokenBinaryIntegerLiteral {
		t.Value = []string{"0b0", "0b1"}[ts.intn(2)]
	} else if ty == lexer.TokenOctalIntegerLiteral {
		t.Value = []string{"0o0", "0o1"}[ts.intn(2)]
	} else if ty == lexer.TokenDecimalIntegerLiteral {
		t.Value = []string{"000", "-1.0"}[ts.intn(2)]
	} else if ty == lexer.TokenHexadecimalIntegerLiteral {
		t.Value = []string{"0x0", "0x7ffffffffffff"}[ts.intn(2)]
	}
	return t
}
