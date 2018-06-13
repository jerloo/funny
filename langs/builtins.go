package langs

import (
	"fmt"
	"time"
	"encoding/base64"
)

type BuiltinFunction = func(interpreter *Interpreter, args []Value) Value

var (
	FUNCTIONS = map[string]BuiltinFunction{
		"echo":         Echo,
		"now":          Now,
		"base64encode": Base64Encode,
		"base64decode": Base64Decode,
	}
)

// ack check function arguments count valid
func ack(args []Value, count int) {
	if len(args) != count {
		panic(fmt.Sprintf("%d arguments required but got %d", count, len(args)))
	}
}

// Echo builtin function echos one or every item in a array
func Echo(interpreter *Interpreter, args []Value) Value {
	fmt.Sprint(interpreter.Vars)
	for _, item := range args {
		fmt.Print(item)
	}
	return nil
}

// Now builtin function return now time
func Now(interpreter *Interpreter, args []Value) Value {
	return Value(time.Now())
}

// Base64Encode return base64 encoded string
func Base64Encode(interpreter *Interpreter, args []Value) Value {
	base64encode := func(val string) string {
		return base64.StdEncoding.EncodeToString([]byte(val))
	}
	if len(args) == 1 {
		return Value(base64encode(args[0].(string)))
	}
	var results []string
	for _, item := range args {
		results = append(results, base64encode(item.(string)))
	}
	return Value(results)
}

// Base64Decode return base64 decoded string
func Base64Decode(interpreter *Interpreter, args []Value) Value {
	base64decode := func(val string) string {
		sb, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			panic(err)
		}
		return string(sb)
	}
	if len(args) == 1 {
		return Value(base64decode(args[0].(string)))
	}
	var results []string
	for _, item := range args {
		results = append(results, base64decode(item.(string)))
	}
	return Value(results)
}
