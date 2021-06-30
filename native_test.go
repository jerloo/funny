package funny

import "testing"

func TestTyping(t *testing.T) {
	d := Typing(&Token{
		Data: "hello",
	})
	if d != "*langs.Token" {
		t.Error(d)
	} else {
		t.Log(d)
	}
}
