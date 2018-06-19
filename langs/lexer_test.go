package langs

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

const (
	DATA = "a = 1\nb=2\nc= a + b"
)

var (
	lexer = NewLexer([]byte(DATA))
)

func TestLexer_LA(t *testing.T) {
	assert.Equalf(t, "a", string(lexer.LA(1)), "")
	assert.Equalf(t, " ", string(lexer.LA(2)), "")
	assert.Equalf(t, "=", string(lexer.LA(3)), "")
	assert.Equalf(t, " ", string(lexer.LA(4)), "")
}

func TestLexer_Consume(t *testing.T) {
	assert.Equalf(t, "a", string(lexer.Consume(1)), "")
	assert.Equalf(t, " ", string(lexer.Consume(1)), "")
	assert.Equalf(t, " ", string(lexer.Consume(2)), "")
}
