package langs

import "testing"

func TestTyping(t *testing.T) {
	d := Typing(&Token{
		Data: "hello",
	})
	if d != "hello" {
		t.Error(d)
	} else {
		t.Log(d)
	}
}
