/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
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

package lexer

import "fmt"

type TokenStream interface {
	// Next consumes and returns one Token. If there are no tokens remaining, it returns Token{TokenEOF}
	Next() Token
	// Input returns the whole input as source code
	Input() string
}

type SeekableTokenStream interface {
	TokenStream
	Cursor() int
	Revert(cursor int)
}

func NewSeekableTokenStream(delegate TokenStream) SeekableTokenStream {
	if s, ok := delegate.(SeekableTokenStream); ok {
		return s
	}
	return &delegatingSeekableTokenStream{delegate: delegate}
}

type seekableTokenStream struct {
	// the offset in the token stream
	cursor int
	// the tokens of the stream
	tokens []Token
}

func (l *seekableTokenStream) Next() Token {
	if l.cursor < len(l.tokens) {
		t := l.tokens[l.cursor]
		if t.Type != TokenEOF {
			l.cursor++
		}
		return t
	}
	panic("unimplemented")
}

func (l *seekableTokenStream) Cursor() int {
	return l.cursor
}

func (l *seekableTokenStream) Revert(cursor int) {
	if cursor > l.cursor {
		panic(fmt.Errorf("illegal forward revert %d, %d", cursor, l.cursor))
	}
	l.cursor = cursor
}

type delegatingSeekableTokenStream struct {
	seekableTokenStream
	delegate TokenStream
}

func (d *delegatingSeekableTokenStream) Next() Token {
	if d.cursor >= len(d.tokens) {
		d.tokens = append(d.tokens, d.delegate.Next())
	}
	return d.seekableTokenStream.Next()
}

func (d *delegatingSeekableTokenStream) Input() string {
	return d.delegate.Input()
}
