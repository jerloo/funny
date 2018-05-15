package cores

import (
	"fmt"
	"reflect"
)

func Typing(data interface{}) string {
	t := reflect.TypeOf(data)
	return fmt.Sprintf("T: %s", t.String())
}
