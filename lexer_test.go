package funny

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DATA = "a = 1\nb=2\nc= a + b"
)

var cases = []Token{
	{
		Kind: NAME,
		Position: Position{
			Line: 0,
			Col:  0,
		},
	},
	{
		Kind: EQ,
		Position: Position{
			Line: 0,
			Col:  2,
		},
	},
	{
		Kind: INT,
		Position: Position{
			Line: 0,
			Col:  4,
		},
	},
	{
		Kind: NEW_LINE,
		Position: Position{
			Line: 0,
			Col:  5,
		},
	},
	{
		Kind: NAME,
		Position: Position{
			Line: 1,
			Col:  0,
		},
	},
	{
		Kind: EQ,
		Position: Position{
			Line: 1,
			Col:  1,
		},
	},
	{
		Kind: INT,
		Position: Position{
			Line: 1,
			Col:  2,
		},
	},
	{
		Kind: NEW_LINE,
		Position: Position{
			Line: 1,
			Col:  3,
		},
	},
	{
		Kind: NAME,
		Position: Position{
			Line: 2,
			Col:  0,
		},
	},
	{
		Kind: EQ,
		Position: Position{
			Line: 2,
			Col:  1,
		},
	},
	{
		Kind: NAME,
		Position: Position{
			Line: 2,
			Col:  3,
		},
	},
	{
		Kind: PLUS,
		Position: Position{
			Line: 2,
			Col:  5,
		},
	},
	{
		Kind: NAME,
		Position: Position{
			Line: 2,
			Col:  7,
		},
	},
	{
		Kind: EOF,
		Position: Position{
			Line: 2,
			Col:  7,
		},
	},
}

func TestLexer_LA(t *testing.T) {
	lexer := NewLexer([]byte(DATA), "")
	assert.Equalf(t, "a", string(lexer.LA(1)), "")
	assert.Equalf(t, " ", string(lexer.LA(2)), "")
	assert.Equalf(t, "=", string(lexer.LA(3)), "")
	assert.Equalf(t, " ", string(lexer.LA(4)), "")
}

func TestLexer_Consume(t *testing.T) {
	lexer := NewLexer([]byte(DATA), "")
	assert.Equalf(t, "a", string(lexer.Consume(1)), "")
	assert.Equalf(t, " ", string(lexer.Consume(1)), "")
	assert.Equalf(t, " ", string(lexer.Consume(2)), "")
}

func TestLexer_Next(t *testing.T) {
	lexer := NewLexer([]byte(DATA), "")
	assert.Equal(t, NAME, lexer.Next().Kind)
	assert.Equal(t, EQ, lexer.Next().Kind)
	assert.Equal(t, INT, lexer.Next().Kind)
	assert.Equal(t, NEW_LINE, lexer.Next().Kind)
	assert.Equal(t, NAME, lexer.Next().Kind)
	assert.Equal(t, EQ, lexer.Next().Kind)
}

func TestLexer_Position(t *testing.T) {
	lexer := NewLexer([]byte(DATA), "")
	tokens := make([]Token, 0)
	for {
		token := lexer.Next()
		tokens = append(tokens, token)
		if token.Kind == EOF {
			break
		}
	}
	assert.Equal(t, len(cases), len(tokens))
	for index, actual := range tokens {
		expect := cases[index]
		assert.Equal(t, expect.Position.Line, actual.Position.Line, actual.String())
		assert.Equal(t, expect.Position.Col, actual.Position.Col, actual.String())
	}
}
