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
	block, err := parser.Parse()
	if err != nil {
		panic(err)
	}
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
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
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
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
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
        } else {
        }
    }
    `), "")
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	fmt.Println(items.Statements[0].String())
}

func TestParseInExpression(t *testing.T) {
	parser := NewParser([]byte(`a = b in [2]`), "")
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseNotInExpression(t *testing.T) {
	parser := NewParser([]byte(`a = 2 not in [2,3]`), "")
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
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
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseIfIn(t *testing.T) {
	parser := NewParser([]byte(`
if 1 in [1,2] {
  minusAccount = 'Assets:Alipay:Balance'
  plusAccount = 'Assets:Others'
}
`), "")
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}

func TestParseIfInStrArray(t *testing.T) {
	parser := NewParser([]byte(`
if direction == '支出' {
  if status in ['交易成功','支付成功','代付成功','亲情卡付款成功','等待确认收货','等待对方发货','交易关闭','充值成功','已付款'] {
    minusAccount = ''
    plusAccount = ''
  }
}
`), "")
	items, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	echoJson, err := prettyjson.Marshal(items)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(echoJson))
}
