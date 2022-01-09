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
	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/cadence/runtime/parser2/lexer"
)

type RandomTokenStream struct {
	Fuzzbits    Fuzzbits
	sentKeyword bool
}

func (ts *RandomTokenStream) Input() string {
	panic("unimplemented")
}

func (ts *RandomTokenStream) Next() lexer.Token {

	if ts.Fuzzbits.BitsLeft() <= 0 { // reached end of random fuzz bits?
		if ts.Fuzzbits.Intn(3) == 1 { // ... then prefer EOF now
			return lexer.Token{Type: lexer.TokenEOF}
		}
	}

	if !ts.sentKeyword {
		ts.sentKeyword = true
		firstWhat := ts.Fuzzbits.Intn(4)
		t := lexer.Token{Type: lexer.TokenIdentifier}
		if firstWhat <= 1 { // 50% of the time, keyword
			t.Value = parser2.Keywords[ts.Fuzzbits.Intn(len(parser2.Keywords))]
		} else if firstWhat == 2 { // 25%, start with some other identifier
			ids := exampleTokenValues[lexer.TokenIdentifier]
			t.Value = ids[ts.Fuzzbits.Intn(len(ids))]
		} else if firstWhat == 3 { // 25%, start with a pragma
			t.Type = lexer.TokenPragma
		}
		return t
	}

	min := int(lexer.TokenBinaryIntegerLiteral)
	max := int(lexer.TokenPragma)
	ty := lexer.TokenType(min + ts.Fuzzbits.Intn(max-min+1))

	t := lexer.Token{Type: ty}
	if ty >= lexer.TokenBinaryIntegerLiteral && ty <= lexer.TokenString {
		egs := exampleTokenValues[ty]
		t.Value = egs[ts.Fuzzbits.Intn(len(egs))]
	}
	return t
}

//////////////////////////////////////////////////////////

type SimpleRandomTokenStream struct {
	Fuzzbits
}

func (s *SimpleRandomTokenStream) Input() string {
	panic("unimplemented")
}

func dt(tokenType lexer.TokenType, value interface{}) lexer.Token {
	return lexer.Token{Type: tokenType, Value: value}
}

// static "assert len(parser2.Keywords) == 43"
var kz = len(parser2.Keywords) - 43

var simpleRandomTokenList = [90]lexer.Token{
	dt(lexer.TokenBinaryIntegerLiteral, "0b0"),
	dt(lexer.TokenOctalIntegerLiteral, "0o0"),
	dt(lexer.TokenDecimalIntegerLiteral, "0"),
	dt(lexer.TokenHexadecimalIntegerLiteral, "0x0"),
	dt(lexer.TokenFixedPointNumberLiteral, "0.0"),
	dt(lexer.TokenPlus, nil),
	dt(lexer.TokenMinus, nil),
	dt(lexer.TokenStar, nil),
	dt(lexer.TokenSlash, nil),
	dt(lexer.TokenPercent, nil),
	dt(lexer.TokenDoubleQuestionMark, nil),
	dt(lexer.TokenParenOpen, nil),
	dt(lexer.TokenParenClose, nil),
	dt(lexer.TokenBraceOpen, nil),
	dt(lexer.TokenBraceClose, nil),
	dt(lexer.TokenBracketOpen, nil),
	dt(lexer.TokenBracketClose, nil),
	dt(lexer.TokenQuestionMark, nil),
	dt(lexer.TokenQuestionMarkDot, nil),
	dt(lexer.TokenComma, nil),
	dt(lexer.TokenColon, nil),
	dt(lexer.TokenDot, nil),
	dt(lexer.TokenSemicolon, nil),
	dt(lexer.TokenLeftArrow, nil),
	dt(lexer.TokenLeftArrowExclamation, nil),
	dt(lexer.TokenSwap, nil),
	dt(lexer.TokenLess, nil),
	dt(lexer.TokenLessEqual, nil),
	dt(lexer.TokenLessLess, nil),
	dt(lexer.TokenGreater, nil),
	dt(lexer.TokenGreaterEqual, nil),
	dt(lexer.TokenEqual, nil),
	dt(lexer.TokenEqualEqual, nil),
	dt(lexer.TokenExclamationMark, nil),
	dt(lexer.TokenNotEqual, nil),
	dt(lexer.TokenAmpersand, nil),
	dt(lexer.TokenAmpersandAmpersand, nil),
	dt(lexer.TokenCaret, nil),
	dt(lexer.TokenVerticalBar, nil),
	dt(lexer.TokenVerticalBarVerticalBar, nil),
	dt(lexer.TokenAt, nil),
	dt(lexer.TokenAsExclamationMark, nil),
	dt(lexer.TokenAsQuestionMark, nil),
	dt(lexer.TokenPragma, nil),
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+0]),  //keywordIf,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+1]),  //keywordElse,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+2]),  //keywordWhile,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+3]),  //keywordBreak,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+4]),  //keywordContinue,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+5]),  //keywordReturn,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+6]),  //keywordTrue,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+7]),  //keywordFalse,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+8]),  //keywordNil,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+9]),  //keywordLet,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+10]), //keywordVar,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+11]), //keywordFun,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+12]), //keywordAs,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+13]), //keywordCreate,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+14]), //keywordDestroy,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+15]), //keywordFor,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+16]), //keywordIn,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+17]), //keywordEmit,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+18]), //keywordAuth,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+19]), //keywordPriv,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+20]), //keywordPub,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+21]), //keywordAccess,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+22]), //keywordSet,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+23]), //keywordAll,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+24]), //keywordSelf,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+25]), //keywordInit,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+26]), //keywordContract,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+27]), //keywordAccount,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+28]), //keywordImport,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+29]), //keywordFrom,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+30]), //keywordPre,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+31]), //keywordPost,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+32]), //keywordEvent,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+33]), //keywordStruct,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+34]), //keywordResource,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+35]), //keywordInterface,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+36]), //KeywordTransaction,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+37]), //keywordPrepare,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+38]), //keywordExecute,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+39]), //keywordCase,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+40]), //keywordSwitch,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+41]), //keywordDefault,
	dt(lexer.TokenIdentifier, parser2.Keywords[kz+42]), //keywordEnum,
	dt(lexer.TokenString, "\"a\""),
	dt(lexer.TokenIdentifier, "a"),
	dt(lexer.TokenIdentifier, "b"),
}

func (s *SimpleRandomTokenStream) Next() lexer.Token {
	if s.Fuzzbits.BitsLeft() <= 0 {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return simpleRandomTokenList[s.Fuzzbits.Intn(len(simpleRandomTokenList))]
}
