package funny

import (
	"fmt"
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
			Line: 1,
			Col:  1,
		},
	},
	{
		position: Position{
			Line: 1,
			Col:  6,
		},
	},
	{
		position: Position{
			Line: 2,
			Col:  1,
		},
	},
	{
		position: Position{
			Line: 2,
			Col:  4,
		},
	},
	{
		position: Position{
			Line: 3,
			Col:  1,
		},
	},
}

func TestParserPosition(t *testing.T) {
	assert.Equal(t, 1, 1)

	parser := NewParser([]byte(ParserTestData))
	blocks := parser.Parse()
	for index, item := range blocks {
		fmt.Printf("%d %s %s\n", index, Typing(item), item.String())
		assert.Equal(t, statements[index].position.Line, item.Position().Line, "Line: "+item.String())
		assert.Equal(t, statements[index].position.Col, item.Position().Col, "Col: "+item.String())
	}
}
