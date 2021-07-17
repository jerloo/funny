package funny

import (
	"testing"

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
