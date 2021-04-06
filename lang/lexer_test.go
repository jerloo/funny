package lang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DATA = "a = 1\nb=2\nc= a + b"
)

func TestLexer_LA(t *testing.T) {
	lexer := NewLexer([]byte(DATA))
	assert.Equalf(t, "a", string(lexer.LA(1)), "")
	assert.Equalf(t, " ", string(lexer.LA(2)), "")
	assert.Equalf(t, "=", string(lexer.LA(3)), "")
	assert.Equalf(t, " ", string(lexer.LA(4)), "")
}

func TestLexer_Consume(t *testing.T) {
	lexer := NewLexer([]byte(DATA))
	assert.Equalf(t, "a", string(lexer.Consume(1)), "")
	assert.Equalf(t, " ", string(lexer.Consume(1)), "")
	assert.Equalf(t, " ", string(lexer.Consume(2)), "")
}

func TestLexer_Next(t *testing.T) {
	lexer := NewLexer([]byte(DATA))
	assert.Equal(t, NAME, lexer.Next().Kind)
	assert.Equal(t, EQ, lexer.Next().Kind)
	assert.Equal(t, INT, lexer.Next().Kind)
	assert.Equal(t, NEW_LINE, lexer.Next().Kind)
	assert.Equal(t, NAME, lexer.Next().Kind)
	assert.Equal(t, EQ, lexer.Next().Kind)
}
