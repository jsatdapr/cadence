package lexer

import (
	"testing"
)

type CannedTokenStream struct {
	Tokens []Token
	offset int
}

var _ TokenStream = &CannedTokenStream{}

func (c *CannedTokenStream) Next() Token {
	if c.offset >= len(c.Tokens) {
		return Token{Type: TokenEOF}
	}
	t := c.Tokens[c.offset]
	c.offset++
	return t
}

func (c *CannedTokenStream) Input() string {
	panic("unimplemented")
}

func TestAutospacingTokenStream(t *testing.T) {
	asts := AutoSpacingTokenStream{
		Delegate: &CannedTokenStream{Tokens: []Token{
			{Type: TokenPragma},
			{Type: TokenIdentifier, Value: "a"},
			{Type: TokenBinaryIntegerLiteral, Value: "0b0"},
			{Type: TokenDecimalIntegerLiteral, Value: "0"},
		}},
	}
	if !(asts.Next().Is(TokenPragma) &&
		asts.Next().Is(TokenIdentifier) &&
		asts.Next().Is(TokenSpace) &&
		asts.Next().Is(TokenBinaryIntegerLiteral) &&
		asts.Next().Is(TokenSpace) &&
		asts.Next().Is(TokenDecimalIntegerLiteral) &&
		asts.Next().Is(TokenEOF) &&
		asts.Next().Is(TokenEOF)) {
		t.FailNow()
	}
}
