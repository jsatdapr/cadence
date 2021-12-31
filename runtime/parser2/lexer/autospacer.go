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

package lexer

import "fmt"

// AutoSpacingTokenStream wraps another TokenStream,
// and inserts spaces where absolutely necessary
type AutoSpacingTokenStream struct {
	Delegate TokenStream
	prevType TokenType
	next     Token
	haveNext bool
}

func (s *AutoSpacingTokenStream) Input() string {
	return s.Delegate.Input()
}

func (s *AutoSpacingTokenStream) Next() (token Token) {
	if s.haveNext {
		token = s.next
		s.haveNext = false
	} else {
		token = s.Delegate.Next()
	}

	if token.Type != TokenEOF && token.Type != TokenSpace {
		if tokenPairsThatCanNeverOccur[s.prevType][token.Type] {
			panic(fmt.Errorf("AutoSpacingTokenStream noticed impossible token pair: %s, %s\n",
				s.prevType.Name(), token.Type.Name()))
		}
		if tokenPairsThatNeedSpacing[s.prevType][token.Type] {
			s.haveNext = true
			s.next = token
			if s.prevType == TokenLineComment {
				token = Token{Type: TokenSpace, Value: Space{String: "\n", ContainsNewline: true}}
			} else {
				token = Token{Type: TokenSpace, Value: Space{String: " ", ContainsNewline: false}}
			}
		}
	}
	s.prevType = token.Type
	return token
}
