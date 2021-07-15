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
