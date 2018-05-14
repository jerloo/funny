package cores

import "fmt"

type Value interface {
}

type Scope map[string]Value

type Interpreter struct {
	Vars []Scope
}

func NewInterpreter(vars Scope) *Interpreter {
	return &Interpreter{
		Vars: []Scope{vars},
	}
}

func (i *Interpreter) Run(program Program) {
	for _, item := range program.Statements {
		switch item := item.(type) {
		case *Assign:
			switch a := item.Target.(type) {
			case *Variable:
				i.Assign(a.Name, i.EvalExpression(item.Value))
			case *Field:
				i.AssignField(a, i.EvalExpression(item.Value))
			}
		case *IFStatement:
		case *FORStatement:
		case *FunctionCall:
			var params []Value
			for _, p := range item.Parameters {
				params = append(params, i.EvalExpression(p))
			}
			if fn, ok := FUNCTIONS[item.Name]; ok {
				fn(i, params)
			}
		case *Function:
		default:
			panic("")
		}
	}
}

func (i *Interpreter) AssignField(field *Field, val Value) {

}

func (i *Interpreter) Assign(name string, val Value) {
	i.Vars[len(i.Vars)-1][name] = val
}

func (i *Interpreter) Lookup(name string) Value {
	for _, item := range i.Vars {
		for k, v := range item {
			if k == name {
				return v
			}
		}
	}
	return nil
}

func (i *Interpreter) PopScope() {
	i.Vars = i.Vars[:len(i.Vars)-1]
}

func (i *Interpreter) PushScope(scope Scope) {
	i.Vars = append(i.Vars, scope)
}

func (i *Interpreter) EvalExpression(expression Expresion) Value {
	switch item := expression.(type) {
	case *BinaryExpression:
		switch item.Operator.Kind {
		case PLUS:
			i.EvalPlus(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case MINUS:
			i.EvalMinus(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case TIMES:
			i.EvalTimes(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case DEVIDE:
			i.EvalDevide(i.EvalExpression(item.Left), i.EvalExpression(item.Right))

		}
	case *List:
	case *Block: // dict => map[string]Value{}
		var scope Scope
		return scope
	case *Boolen:

	}
	panic(fmt.Sprintf("eval expression error : %s", expression.String()))
}

func (i *Interpreter) EvalPlus(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left + right)
		}
	case *[]Value:
		if right, ok := right.(*[]Value); ok {
			s := make([]Value, 0, len(*left)+len(*right))
			s = append(s, *left...)
			s = append(s, *right...)
			return Value(&s)
		}
	case *Scope:
		var s []Value
		if right, ok := right.(*Scope); ok {
			for _, l := range *left {
				flag := false

				for _, r := range s {
					if !i.EvalEqual(l, r).(bool) {
						flag = true
					} else {
						flag = false
					}

				}
				if !flag {
					s = append(s, l)
				}
			}
			for _, r := range *right {
				flag := false
				for _, c := range s {
					if !i.EvalEqual(r, c).(bool) {
						flag = true
					} else {
						flag = false
					}
				}
				if !flag {
					s = append(s, r)
				}
			}
		}
		return s
	}
	panic("eval plus requires types: int, list, dict")
}

func (i *Interpreter) EvalMinus(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left - right)
		}
	case *[]Value:
		var s []Value
		if right, ok := right.(*Scope); ok {
			for _, l := range *left {
				for _, r := range *right {
					if i.EvalEqual(l, r).(bool) {
						s = append(s, l)
					}
				}
			}
		}
		return s
	case *Scope:
		var s []Value
		if right, ok := right.(*Scope); ok {
			for _, l := range *left {
				for _, r := range *right {
					if i.EvalEqual(l, r).(bool) {
						s = append(s, l)
					}
				}
			}
		}
		return s
	}
	panic("eval plus requires types: int, list, dict")
}

func (i *Interpreter) EvalTimes(left, right Value) Value {
	if l, ok := left.(int); ok {
		if r, o := right.(int); o {
			return Value(l * r)
		}
	}
	panic("eval plus times types: int")
}

func (i *Interpreter) EvalDevide(left, right Value) Value {
	if l, o := left.(int); o {
		if r, k := right.(int); k {
			return Value(l / r)
		}
	}
	panic("eval plus devide types: int")
}

func (i *Interpreter) EvalEqual(left, right Value) Value {
	switch l := left.(type) {
	case nil:
		return Value(right == nil)
	case int:
		if r, ok := right.(int); ok {
			return Value(l == r)
		}
	case *[]Value:
		if r, ok := right.(*[]Value); ok {
			if len(*l) != len(*r) {
				return Value(false)
			}
			for _, itemL := range *l {
				for _, itemR := range *r {
					if !i.EvalEqual(itemL, itemR).(bool) {
						return Value(false)
					}
				}
			}
			return Value(true)
		}
	case *Scope:
		if r, ok := right.(*Block); ok {
			if len(*l) != len(*r) {
				return Value(false)
			}
			for _, itemL := range *l {
				for _, itemR := range *r {
					if !i.EvalEqual(itemL, itemR).(bool) {
						return Value(false)
					}
				}
			}
			return Value(true)
		}
	}
	return Value(false)
}
