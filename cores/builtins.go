package cores

import "fmt"

type BuiltinFunction = func(interpreter *Interpreter, args []Value) Value

var (
	FUNCTIONS = map[string]BuiltinFunction{
		"echo": echo,
	}
)

func echo(interpreter *Interpreter, args []Value) Value {
	fmt.Sprint(interpreter.Vars)
	for _, item := range args {
		fmt.Print(item)
	}
	return nil
}
