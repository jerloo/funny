package langs

import (
	"testing"
)

func RunSingle(data interface{}) (*Interpreter, Value) {
	i := NewInterpreterWithScope(make(map[string]Value))
	var d []byte
	switch data.(type) {
	case string:
		d = []byte(data.(string))
	case []byte:
	}
	parser := NewParser(d)
	r := i.Run(Program{
		Statements: parser.Parse(),
	})
	return i, Value(r)
}

func TestInterpreter_Assign(t *testing.T) {
	i := NewInterpreterWithScope(make(map[string]Value))
	i.Assign("a", Value(1))
	flag := false
	var val interface{}
	for _, scope := range i.Vars {
		for k, v := range scope {
			if k == "a" {
				flag = true
				val = v
			}
		}
	}
	if !flag {
		t.Error("assign error key not in scope")
	} else {
		if val != 1 {
			t.Error("assign error value not equal 1")
		}
	}
}

func TestInterpreter_Lookup(t *testing.T) {
	i := NewInterpreterWithScope(make(map[string]Value))
	i.Assign("a", Value(1))
	val := i.Lookup("a")
	if val != 1 {
		t.Error("lookup error")
	}
}

func TestInterpreter_EvalFunctionCall(t *testing.T) {
	i := NewInterpreterWithScope(make(map[string]Value))
	parser := NewParser([]byte("echo(1)"))
	i.Run(Program{
		parser.Parse(),
	})
}

func TestInterpreter_EvalFunctionCall2(t *testing.T) {
	i := NewInterpreterWithScope(make(map[string]Value))
	parser := NewParser([]byte("echo2(b){echo(b)} \n echo2(1)"))
	i.Run(Program{
		parser.Parse(),
	})
}

func TestInterpreter_EvalPlus(t *testing.T) {
	i := NewInterpreterWithScope(make(map[string]Value))
	parser := NewParser([]byte("  a = 1 + 1"))
	i.Run(Program{
		parser.Parse(),
	})
	a := i.Lookup("a")
	if a != 2 {
		t.Error("eval plus error")
	}
}

func TestInterpreter_Run(t *testing.T) {
	data := `
a = 1
b = 2
c = a + b

echo(c)

p(a, b){
    return a + b
}

d = p(a,b)

return d - 1`

	_, r := RunSingle(data)
	if r != 2 {
		t.Error("RunSingle funny.fun must return 2")
	}
}
