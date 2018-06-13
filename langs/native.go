package langs

import (
	"fmt"
	"reflect"
)

func Typing(data interface{}) string {
	t := reflect.TypeOf(data)
	return fmt.Sprintf("%s", t.String())
}
