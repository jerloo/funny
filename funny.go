package funny

import (
	"fmt"
	"os"
	"path"
)

// Value one value of some like variable
type Value interface {
}

// Scope stores variables
type Scope map[string]Value

// Funny the virtual machine of funny code
type Funny struct {
	Vars      []Scope
	Functions map[string]BuiltinFunction

	Current Position
}

// NewFunnyWithScope create a new funny
func NewFunnyWithScope(vars Scope) *Funny {
	return &Funny{
		Vars: []Scope{
			vars,
		},
		Functions: FUNCTIONS,
	}
}

// Create a new funny with default settings
func NewFunny() *Funny {
	return &Funny{
		Vars: []Scope{
			make(map[string]Value),
		},
		Functions: FUNCTIONS,
	}
}

// Debug get debug value
func (i *Funny) Debug() bool {
	v := i.LookupDefault("debug", Value(false))
	if v == nil {
		return false
	}
	if v, ok := v.(bool); ok {
		return v
	}
	return false
}

func (i *Funny) RunFile(filename string) (Value, bool) {
	if !path.IsAbs(filename) {
		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		filename = path.Join(currentDir, filename)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	parser := NewParser(data, filename)
	parser.ContentFile = filename
	statements, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	program := Program{
		Statements: statements,
	}
	return i.Run(program)
}

// Run the part of the code
func (i *Funny) Run(v interface{}) (Value, bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	switch v := v.(type) {
	case Statement:
		return i.EvalStatement(v)
	case Program:
		return i.Run(&v)
	case *Program:
		return i.EvalBlock(v.Statements)
	case string:
		return i.Run([]byte(v))
	case []byte:
		parser := NewParser(v, "")
		statements, err := parser.Parse()
		if err != nil {
			panic(err)
		}
		program := Program{
			Statements: statements,
		}
		return i.Run(program)
	default:
		panic(P(fmt.Sprintf("unknow type of running value: [%v]", v), i.Current))
	}
}

// EvalBlock eval a block
func (i *Funny) EvalBlock(block *Block) (Value, bool) {
	if block == nil {
		return Value(nil), false
	}
	i.Current = block.GetPosition()
	for _, item := range block.Statements {
		r, has := i.EvalStatement(item)
		if has {
			return r, has
		}
	}
	return Value(nil), false
}

// RegisterFunction register a builtin or customer function
func (i *Funny) RegisterFunction(name string, fn BuiltinFunction) error {
	if _, exists := i.Functions[name]; exists {
		return fmt.Errorf("function [%s] already exists", name)
	}
	i.Functions[name] = fn
	return nil
}

// EvalIfStatement eval if statement
func (i *Funny) EvalIfStatement(item *IFStatement) (Value, bool) {
	i.Current = item.GetPosition()
	exp := i.EvalExpression(item.Condition)
	if exp, ok := exp.(bool); ok {
		if exp {
			r, has := i.EvalBlock(item.Body)
			if has {
				return r, true
			}
		} else if item.ElseIf != nil {
			r, has := i.EvalIfStatement(item.ElseIf.(*IFStatement))
			if has {
				return r, true
			}
		} else {
			r, has := i.EvalBlock(item.Else)
			if has {
				return r, true
			}
		}
	} else {
		panic(P("if statement condition must be boolen value", item.Position))
	}
	return Value(nil), false
}

// EvalForStatement eval for statement
func (i *Funny) EvalForStatement(item *FORStatement) (Value, bool) {
	i.Current = item.GetPosition()
	panic(P("NOT IMPLEMENT", i.Current))
}

// EvalStatement eval statement
func (i *Funny) EvalStatement(item Statement) (Value, bool) {
	i.Current = item.GetPosition()
	switch item := item.(type) {
	case *Assign:
		switch a := item.Target.(type) {
		case *Variable:
			i.Assign(a.Name, i.EvalExpression(item.Value))
		case *Field:
			i.AssignField(a, i.EvalExpression(item.Value))
		default:
			panic(P("invalid assignment", item.Position))
		}
	case *IFStatement:
		val, has := i.EvalIfStatement(item)
		if has {
			return val, true
		}
	case *FORStatement:
		val, has := i.EvalForStatement(item)
		if has {
			return val, true
		}
	case *FunctionCall:
		i.EvalFunctionCall(item)
	case *ImportFunctionCall:
		for _, d := range item.Block.Statements {
			switch d := d.(type) {
			case *Assign:
				if t, ok := d.Target.(*Variable); ok {
					i.Assign(t.Name, i.EvalExpression(d.Value))
				} else {
					panic(P("block assignments must be variable", item.Position))
				}
			case *NewLine:
				break
			case *Comment:
				break
			case *Function:
				i.Assign(d.Name, d)
			default:
				panic(P("module must only contains assignment and func", item.Position))
			}
		}
	case *Return:
		return i.EvalExpression(item.Value), true
	case *Function:
		i.Assign(item.Name, item)
	case *Field:
		i.EvalField(item)
	case *NewLine:
		break
	case *Comment:
		break
	default:
		panic(P(fmt.Sprintf("invalid statement [%s]", item.String()), item.GetPosition()))
	}
	return Value(nil), false
}

// EvalFunctionCall eval function call like test(a, b)
func (i *Funny) EvalFunctionCall(item *FunctionCall) (Value, bool) {
	i.Current = item.GetPosition()
	var params []Value
	for _, p := range item.Parameters {
		params = append(params, i.EvalExpression(p))
	}
	if fn, ok := i.Functions[item.Name]; ok {
		return fn(i, params), true
	}
	this := i.LookupDefault("this", nil)
	var look Value
	if this != nil {
		look = this.(map[string]Value)[item.Name]
	}
	if look == nil {
		look := i.LookupDefault(item.Name, nil)
		if look == nil {
			panic(P(fmt.Sprintf("function [%s] not defined", item.Name), i.Current))
		}
		fun := i.Lookup(item.Name).(*Function)
		return i.EvalFunction(*fun, params)

	} else {
		fun := look.(*Function)
		return i.EvalFunction(*fun, params)
	}
}

// EvalFunction eval function
func (i *Funny) EvalFunction(item Function, params []Value) (Value, bool) {
	i.Current = item.GetPosition()
	if len(params) < len(item.Parameters) {
		panic(P(fmt.Sprintf("function %s required %d args but %d given", item.Name, len(item.Parameters), len(params)), item.Position))
	}
	scope := Scope{}
	i.PushScope(scope)
	for index, p := range item.Parameters {
		i.Assign(p.(*Variable).Name, params[index])
	}
	r, has := i.EvalBlock(item.Body)
	i.PopScope()
	return r, has
}

// AssignField assign one field value
func (i *Funny) AssignField(field *Field, val Value) {
	i.Current = field.GetPosition()
	scope := make(map[string]Value)

	find := i.Lookup(field.Variable.Name)
	if find != nil {
		scope = find.(map[string]Value)
	}
	scope[field.Value.(*Variable).Name] = val
	i.Assign(field.Variable.Name, Value(scope))
}

// Assign assign one variable
func (i *Funny) Assign(name string, val Value) {
	i.Vars[len(i.Vars)-1][name] = val
}

// LookupDefault find one variable named name and get value, if not found, return default
func (i *Funny) LookupDefault(name string, defaultVal Value) Value {
	for index := len(i.Vars) - 1; index >= 0; index-- {
		item := i.Vars[index]
		for k, v := range item {
			if k == name {
				return v
			}
		}
	}
	return defaultVal
}

// Lookup find one variable named name and get value
func (i *Funny) Lookup(name string) Value {
	for index := len(i.Vars) - 1; index >= 0; index-- {
		item := i.Vars[index]
		for k, v := range item {
			if k == name {
				return v
			}
		}
	}
	return Value(nil)
}

// PopScope pop current scope
func (i *Funny) PopScope() {
	i.Vars = i.Vars[:len(i.Vars)-1]
}

// PushScope push scope into current
func (i *Funny) PushScope(scope Scope) {
	i.Vars = append(i.Vars, scope)
}

// EvalExpression eval part that is expression
func (i *Funny) EvalExpression(expression Statement) Value {
	i.Current = expression.GetPosition()
	switch item := expression.(type) {
	case *BinaryExpression:
		// TODO:
		// a is int
		// a is nil
		// a is not int
		// a in [1,2]
		// a not in [1,2]
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
			return i.EvalEqual(i.EvalExpression(item.Left), i.EvalExpression(item.Right))
		case NOTEQ:
			return Value((!i.EvalEqual(i.EvalExpression(item.Left), i.EvalExpression(item.Right)).(bool)))
		case NAME:
			switch item.Operator.Data {
			case NOT:
				switch v := item.Right.(type) {
				case *BinaryExpression:
					if v.Operator.Data == IN {
						return Value(!i.EvalIn(i.EvalExpression(item.Left), v.Right).(bool))
					}
				}
				return Value(!i.EvalExpression(item.Right).(bool))
			case IN:
				return i.EvalIn(i.EvalExpression(item.Left), item.Right)
			}
		default:
			panic(P(fmt.Sprintf("only support [+] [-] [*] [/] [>] [>=] [==] [<=] [<] [in] [not] given [%s]", expression.(*BinaryExpression).Operator.Data), expression.GetPosition()))
		}
	case *List:
		var ls []interface{}
		for _, item := range item.Values {
			ls = append(ls, i.EvalExpression(item))
		}
		return Value(ls)
	case *Block: // dict => map[string]Value{}
		scope := make(map[string]Value)

		for _, d := range item.Statements {
			switch d := d.(type) {
			case *Assign:
				if t, ok := d.Target.(*Variable); ok {
					scope[t.Name] = i.EvalExpression(d.Value)
				} else {
					panic(P("block assignments must be variable", item.Position))
				}
			case *NewLine:
				break
			case *Comment:
				break
			case *Function:
				scope[d.Name] = d
			default:
				panic(P("dict struct must only contains assignment and func", item.Position))
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
		r, _ := i.EvalFunctionCall(item)
		return r
	case *Field:
		return i.EvalField(item)
	case *ListAccess:
		ls := i.Lookup(item.List.Name)
		lsEntry := ls.([]interface{})
		val := lsEntry[item.Index]
		return Value(val)
	case *ImportFunctionCall:
		scope := make(map[string]Value)

		for _, d := range item.Block.Statements {
			switch d := d.(type) {
			case *Assign:
				if t, ok := d.Target.(*Variable); ok {
					scope[t.Name] = i.EvalExpression(d.Value)
				} else {
					panic(P("block assignments must be variable", item.Position))
				}
			case *NewLine:
				break
			case *Comment:
				break
			case *Function:
				scope[d.Name] = d
			default:
				panic(P("module must only contains assignment and func", item.Position))
			}
		}
		return scope
	}
	panic(P(fmt.Sprintf("eval expression error: [%s]", expression.String()), expression.GetPosition()))
}

func (i *Funny) EvalIn(leftValue Value, right Statement) Value {
	switch rightValue := right.(type) {
	case *List:
		for _, item := range rightValue.Values {
			v := i.EvalExpression(item)
			if leftValue == v {
				return Value(true)
			}
		}
	}
	return Value(false)
}

// EvalField person.age
func (i *Funny) EvalField(item *Field) Value {
	i.Current = item.GetPosition()
	root := i.Lookup(item.Variable.Name)
	switch v := item.Value.(type) {
	case *FunctionCall:
		this := root.(map[string]Value)
		scope := Scope{
			"this": this,
		}
		for key, val := range this {
			scope[key] = val
		}
		i.PushScope(scope)
		r, _ := i.EvalFunctionCall(v)
		i.PopScope()
		return r
	case *StringExpression:
		if val, ok := root.(map[string]Value); ok {
			return Value(val[v.Value])
		}
		if val, ok := root.(map[string]interface{}); ok {
			return Value(val[v.Value])
		}
	case *Variable:
		key := i.Lookup(v.Name)
		if keyStr, ok := key.(string); ok {
			if val, ok := root.(map[string]Value); ok {
				return Value(val[keyStr])
			}
			if val, ok := root.(map[string]interface{}); ok {
				return Value(val[keyStr])
			}
		} else {
			panic(P(fmt.Sprintf("unknow type field access key %v", key), i.Current))
		}
	case *Field:
		scope := Scope{}
		if vm, ok := root.(map[string]Value); ok {
			for key, val := range vm {
				scope[key] = val
			}
			i.PushScope(scope)
			r := i.EvalField(v)
			i.PopScope()
			return r
		} else if vm, ok := root.(map[string]interface{}); ok {
			for key, val := range vm {
				scope[key] = val
			}
			i.PushScope(scope)
			r := i.EvalField(v)
			i.PopScope()
			return r
		}
		panic(P(fmt.Sprintf("unknow type %v", v), i.Current))
	default:
		panic(P(fmt.Sprintf("unknow type %v", v), i.Current))
	}
	return Value(nil)
}

// EvalPlus +
func (i *Funny) EvalPlus(left, right Value) Value {
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
	panic(P(fmt.Sprintf("eval plus only support types: [int, list, dict] given [%s]", Typing(left)), i.Current))
}

// EvalMinus -
func (i *Funny) EvalMinus(left, right Value) Value {
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
	panic(P("eval plus only support types: [int, list, dict]", i.Current))
}

// EvalTimes *
func (i *Funny) EvalTimes(left, right Value) Value {
	if l, ok := left.(int); ok {
		if r, o := right.(int); o {
			return Value(l * r)
		}
	}
	panic(P("eval plus times only support types: [int]", i.Current))
}

// EvalDevide /
func (i *Funny) EvalDevide(left, right Value) Value {
	if l, o := left.(int); o {
		if r, k := right.(int); k {
			return Value(l / r)
		}
	}
	panic(P("eval plus devide only support types: [int]", i.Current))
}

// EvalEqual ==
func (i *Funny) EvalEqual(left, right Value) Value {
	switch l := left.(type) {
	case nil:
		return Value(right == nil)
	case int:
		if r, ok := right.(int); ok {
			return Value(l == r)
		}
		if r, ok := right.(float64); ok {
			return Value(float64(l) == r)
		}
	case float64:
		if r, ok := right.(float64); ok {
			return Value(l == r)
		}
		if r, ok := right.(int); ok {
			return Value(l == float64(r))
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
			if len(*l) != len(r.Statements) {
				return Value(false)
			}
			for _, itemL := range *l {
				for _, itemR := range r.Statements {
					if !i.EvalEqual(itemL, itemR).(bool) {
						return Value(false)
					}
				}
			}
			return Value(true)
		}
	case string:
		if r, ok := right.(string); ok {
			return Value(l == r)
		}
		return Value(false)
	default:
		panic(P(fmt.Sprintf("unsupport type [%s]", Typing(l)), i.Current))
	}
	return Value(false)
}

// EvalGt >
func (i *Funny) EvalGt(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left > right)
		}
	}
	panic(P("eval gt only support: [int]", i.Current))
}

// EvalGte >=
func (i *Funny) EvalGte(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left >= right)
		}
	}
	panic(P("eval lte only support: [int]", i.Current))
}

// EvalLt <
func (i *Funny) EvalLt(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left < right)
		}
	}
	panic(P("eval lt only support: [int]", i.Current))
}

// EvalLte <=
func (i *Funny) EvalLte(left, right Value) Value {
	switch left := left.(type) {
	case int:
		if right, ok := right.(int); ok {
			return Value(left <= right)
		}
	}
	panic(P("eval lte only support: [int]", i.Current))
}

// EvalDoubleEq ==
func (i *Funny) EvalDoubleEq(left, right Value) Value {
	return left == right
	// switch left := left.(type) {
	// case int:
	// 	if right, ok := right.(int); ok {
	// 		return Value(left == right)
	// 	}
	// case nil:
	// 	if left == nil && right == nil {
	// 		return Value(true)
	// 	}
	// default:
	// 	return Value(left == right)
	// }
	// panic(P("eval double eq only support: [int]")
}
