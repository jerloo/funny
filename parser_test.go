package funny

import (
	"fmt"
	"testing"

	"github.com/jerloo/go-prettyjson"
	"github.com/stretchr/testify/assert"
)

const (
	ParserTestData = "a = 1\nb=2\nc= a + b"
)

var statements = []struct {
	position Position
}{
	{
		position: Position{
			Line: 0,
			Col:  0,
		},
	},
	{
		position: Position{
			Line: 0,
			Col:  5,
		},
	},
	{
		position: Position{
			Line: 1,
			Col:  0,
		},
	},
	{
		position: Position{
			Line: 1,
			Col:  3,
		},
	},
	{
		position: Position{
			Line: 2,
			Col:  0,
		},
	},
}

func TestNewLineLength(t *testing.T) {
	assert.Equal(t, 1, len("\n"))
}

func TestParserPosition(t *testing.T) {
	assert.Equal(t, 1, 1)

	parser := NewParser([]byte(ParserTestData), "")
	block := parser.Parse()
	for index, item := range block.Statements {
		// fmt.Printf("%d %s %s\n", index, Typing(item), item.String())
		assert.Equal(t, statements[index].position.Line, item.GetPosition().Line, "Line: "+item.String())
		assert.Equal(t, statements[index].position.Col, item.GetPosition().Col, "Col: "+item.String())
	}
}

func TestParseFunctionCall(t *testing.T) {
	parser := NewParser([]byte("echo2(b){echo(b)} \n echo2(1)"), "")
	parser.Parse()
}

func TestParseIfStatement(t *testing.T) {
	parser := NewParser([]byte(`
	a = 1
	if a > 0 {
		echoln(true)
	}
	`), "")
	parser.Parse()
}

func TestParseIfStatementWithElse(t *testing.T) {
	parser := NewParser([]byte(`
a = 1
if a > 0 {
echoln(true)
} else {
echoln('else')
}
`), "")
	items := parser.Parse()
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseIfStatementWithElseIf(t *testing.T) {
	parser := NewParser([]byte(`
a = 1
if a > 0 {
echoln(true)
} else if a == 1 {
echoln('else if')
} else {
echoln('else')
}
`), "")
	items := parser.Parse()
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseIfStatementWithField(t *testing.T) {
	parser := NewParser([]byte(`
	a = {
		t = 1
	}
	if a.t > 0 {
		echoln(true)
	} else if a.t == 1 {
		echoln('else if')
	} else {
		echoln('else')
	}
	`), "")
	parser.Parse()
}

func TestParseIfStatement2(t *testing.T) {
	parser := NewParser([]byte(`
	main(row){
		if a == 1 {
		} else a == 1 {
		}
	}
	`), "")
	blocks := parser.Parse()
	fmt.Println(blocks.Statements[0].String())
}

func TestParseInExpression(t *testing.T) {
	parser := NewParser([]byte(`a = 2 in [2]`), "")
	items := parser.Parse()
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseNotInExpression(t *testing.T) {
	parser := NewParser([]byte(`a = 2 not in [2]`), "")
	items := parser.Parse()
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseIf(t *testing.T) {
	parser := NewParser([]byte(`
if a == 1 {
    b = 2
}`), "")
	items := parser.Parse()
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}
