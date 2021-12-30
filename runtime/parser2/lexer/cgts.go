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

import (
	"bytes"
	"fmt"
	"strings"
)

// CodeGatheringTokenStream wraps another TokenStream,
// remembers its Tokens, and outputs Cadence source code
type CodeGatheringTokenStream struct {
	seekableTokenStream
	Delegate TokenStream
	expected string
	sawError bool
}

func (s *CodeGatheringTokenStream) Next() Token {
	if s.cursor >= len(s.tokens) {
		s.tokens = append(s.tokens, s.Delegate.Next())
	}
	token := s.seekableTokenStream.Next()

	if token.Type == TokenError {
		s.sawError = true
	}

	// on testcases that are expected to succeed...
	if token.Type == TokenEOF && !s.sawError {

		// ... assert that the generated output is as expected
		if s.expected != "" && s.Input() != s.expected {
			panic(fmt.Errorf("\nexpected %q\n     got %q", s.expected, s.Input()))
		}
	}
	return token
}

func (s *CodeGatheringTokenStream) Input() string {
	buffer := bytes.Buffer{}
	for _, t := range s.tokens {
		buffer.WriteString(tokenText(t))
	}
	return buffer.String()
}

func tokenText(t Token) string {
	switch t.Type {
	case TokenSpace:
		return t.Value.(Space).String
	case TokenError, TokenEOF:
		return ""
	}
	if s, ok := t.Value.(string); ok {
		return s
	}
	text := fmt.Sprint(t.Type)
	if text[0] != '\'' {
		panic("unexpected TokenType.text = " + text)
	}
	return strings.Trim(text, "'")
}
