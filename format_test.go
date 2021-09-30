package funny

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testNewLine = `




a(){

}



b(){

}

`
const testNewLineResult = `

a() {

}

b() {

}
`

func TestFormat(t *testing.T) {
	result := Format([]byte(testNewLine), "")
	fmt.Println(result)
	assert.Equal(t, testNewLineResult, result)
}

func TestIfElseFormat(t *testing.T) {
	result := Format([]byte(`
	if a == 1 {
		if b == 2 {
			c = 3
		} else if b == 3 {
			c = 3
		}
	}
	`), "")
	fmt.Println(result)
}

func TestIfElseIfFormat(t *testing.T) {
	result := Format([]byte(`
if a == 1 {
if b == 2 {
c = 3
} else if b == 3 {
c = 3
}
}
`), "")
	fmt.Println(result)
}
