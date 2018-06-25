package langs

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

func (i *Interpreter) Run(program Program) Value {
	for _, item := range program.Statements {
		r := i.EvalStatement(item)
		if r != nil {
			return r
		}
	}
	return Value(nil)
}

func (i *Interpreter) EvalStatement(item Statement) Value {
	switch item := item.(type) {
	case *Assign:
		switch a := item.Target.(type) {
		case *Variable:
			i.Assign(a.Name, i.EvalExpression(item.Value))
		case *Field:
			i.AssignField(a, i.EvalExpression(item.Value))
		}
		break
	case *IFStatement:
	case *FORStatement:
	case *FunctionCall:
		i.EvalFunctionCall(item)
		break
	case *Function:
		i.Assign(item.Name, item)
	case *Field:
		return i.EvalField(item)
	case *Return:
		return i.EvalExpression(item.Value)
	case *NewLine:
		return nil
	default:
		panic(fmt.Sprintf("invalid statement %s", item.String()))
	}
	return Value(nil)
}

func (i *Interpreter) EvalFunctionCall(item *FunctionCall) Value {
	var params []Value
	for _, p := range item.Parameters {
		params = append(params, i.EvalExpression(p))
	}
	if fn, ok := FUNCTIONS[item.Name]; ok {
		return fn(i, params)
	}
	fun := i.Lookup(item.Name).(*Function)
	return i.EvalFunction(*fun, params)
}

func (i *Interpreter) EvalFunction(item Function, params []Value) Value {

	for index, p := range item.Parameters {
		i.Assign(p.(*Variable).Name, params[index])
	}
	for _, b := range item.Body {
		if r := i.EvalStatement(b); r != nil {
			return r
		}
	}
	return Value(nil)
}

func (i *Interpreter) AssignField(field *Field, val Value) {
	scope := make(map[string]Value)
	scope[field.Value.(*Variable).Name] = val
	i.Assign(field.Variable.Name, Value(scope))
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

func (i *Interpreter) EvalExpression(expression Expression) Value {
	switch item := expression.(type) {
	case *BinaryExpression:
		// TODO: string minus
		switch item.Operator.Kind {
		case PLUS:
			return i.EvalPlus(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case MINUS:
			return i.EvalMinus(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case TIMES:
			return i.EvalTimes(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case DEVIDE:
			return i.EvalDevide(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case GT:
			return i.EvalGt(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case GTE:
			return i.EvalGte(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case LT:
			return i.EvalLt(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case LTE:
			return i.EvalLte(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case DOUBLE_EQ:
			return i.EvalDoubleEq(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		default:
			panic("support [+] [-] [*] [/] [>] [>=] [==] [<=] [<]")
		}
	case *List:
		var ls []interface{}
		for _, item := range item.Values {
			ls = append(ls, i.EvalExpression(item))
		}
		return Value(ls)
	case *Block: // dict => map[string]Value{}
		scope := make(map[string]Value)

		for _, d := range *item {
			switch d := d.(type) {
			case *Assign:
				if t, ok := d.Target.(*Variable); ok {
					scope[t.Name] = i.EvalExpression(d.Value)
				} else {
					panic("block assign must variable")
				}
			default:
				panic("block must only contains assign")
			}
		}
		return scope
	case *Boolen:
		return Value(item.Value)
	case *Variable:
		return i.Lookup(item.Name)
	case *Literal:
		return Value(item.Value)
	case *FunctionCall:
		return i.EvalFunctionCall(item)
	case *Field:
		return i.EvalField(item)
	}
	panic(fmt.Sprintf("eval expression error : %s", expression.String()))
}

func (i *Interpreter) EvalField(item *Field) Value {
	switch v := item.Value.(type) {
	case *FunctionCall:
		return i.EvalFunctionCall(v)
	case *Variable:
		ii := i.Lookup(item.Variable.Name).(map[string]Value)
		return Value(ii[v.Name])
	}
	return Value(nil)
}

func (i *Interpreter) EvalPlus(left, right Value) Value {
	switch left := left.(type) {
	case string:
		if right, ok := right.(string); ok {
			return Value(left + right)
		}
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

func (i *Interpreter) EvalGt(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left > right)
		}
	}
	panic("eval gt support: [int]")
}

func (i *Interpreter) EvalGte(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left >= right)
		}
	}
	panic("eval lte support: [int]")
}

func (i *Interpreter) EvalLt(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left < right)
		}
	}
	panic("eval lt support: [int]")
}

func (i *Interpreter) EvalLte(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left <= right)
		}
	}
	panic("eval lte support: [int]")
}

func (i *Interpreter) EvalDoubleEq(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left == right)
		}
	}
	panic("eval double eq support: [int]")
}
