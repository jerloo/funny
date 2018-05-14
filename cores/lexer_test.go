package cores

import (
	"testing"
	"unicode/utf8"
)

func TestLexer_Consume(t *testing.T) {
	s := "abcdefg"
	r, size := utf8.DecodeRune([]byte(s))
	if r != 'a' {
		t.Error(size)
	}
	r, size= utf8.DecodeRune([]byte(s)[size:])
	if r != 'b' {
		t.Error(size)
	}
}
