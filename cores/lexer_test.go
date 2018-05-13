package cores

import (
	"io/ioutil"
	"testing"
)

func TestLexer_LA(t *testing.T) {
	ss := "echo(t.sub)"
	lexer := NewLexer([]byte(ss))
	for {
		token := lexer.Next()
		t.Logf(token.String())
		if token.Kind == EOF {
			break
		}
	}
}

func TestParse(t *testing.T) {
	data, err := ioutil.ReadFile("funny.fl")
	if err != nil {
		panic(err)
	}
	lexer := NewLexer(data)
	for {
		t.Log("fasfasff")
		token := lexer.Next()
		t.Logf("token: %c", token.Data)
		if token.Kind == EOF {
			break
		}
	}

}

func TestToken_String(t *testing.T) {
	token := Token{}
	t.Log(token)
}
