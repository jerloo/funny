package funny

import (
	"reflect"
)

// Typing return the type name of one object
func Typing(data interface{}) string {
	t := reflect.TypeOf(data)
	if t == nil {
		return "nil"
	}
	return t.String()
}
