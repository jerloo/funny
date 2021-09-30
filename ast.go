package funny

import (
	"fmt"
	"strings"
)

func collectBlock(block *Block) []string {
	flag := 0
	var s []string
	for _, item := range block.Statements {
		if item == nil {
			break
		}
		switch item.(type) {
		case *NewLine:
			flag++
			if flag < 1 {
				continue
			}
		}
		flag = 0
		s = append(s, item.String())
	}
	return s
}

func intent(s string) string {
	ss := strings.Split(s, "\n")
	for index, item := range ss {
		if item == "" {
			continue
		}
		ss[index] = fmt.Sprintf("  %s", strings.TrimRight(item, " \n"))
	}
	return strings.Join(ss, "\n")
}

const (
	STNewLine            = "NewLine"
	STVariable           = "Variable"
	STLiteral            = "Literal"
	STBinaryExpression   = "BinaryExpression"
	STSubExpression      = "SubExpression"
	STAssign             = "Assign"
	STBlock              = "Block"
	STList               = "List"
	STListAccess         = "ListAccess"
	STFunction           = "Function"
	STFunctionCall       = "FunctionCall"
	STImportFunctionCall = "Import"
	STIfStatement        = "IfStatement"
	STForStatement       = "ForStatement"
	STIterableExpression = "IterableExpression"
	STBreak              = "Break"
	STContinue           = "Continue"
	STReturn             = "Return"
	STField              = "Field"
	STBoolean            = "Boolean"
	STStringExpression   = "StringExpression"
	STComment            = "Comment"
)

// Statement abstract
type Statement interface {
	String() string
	GetPosition() Position
	// EndPosition() Position
}

// NewLine @impl Statement \n
type NewLine struct {
	Position Position
	Type     string
}

func (n *NewLine) String() string {
	return "\n"
}

func (n *NewLine) GetPosition() Position {
	return n.Position
}

func (n *NewLine) EndPosition() Position {
	return Position{
		File:   n.Position.File,
		Line:   n.Position.Line,
		Col:    n.Position.Col + len(n.String()),
		Length: 0,
	}
}

// Variable means var
type Variable struct {
	Position Position
	Type     string

	Name string
}

func (v *Variable) GetPosition() Position {
	return v.Position
}

func (v *Variable) String() string {
	if strings.Contains(v.Name, "-") {
		return fmt.Sprintf("'%s'", v.Name)
	}
	return v.Name
}

// Literal like 1
type Literal struct {
	Position Position
	Type     string

	Value interface{}
}

func (v *Literal) GetPosition() Position {
	return v.Position
}

func (l *Literal) String() string {
	if Typing(l.Value) == "string" {
		return fmt.Sprintf("'%v'", l.Value)
	}
	return fmt.Sprintf("%v", l.Value)
}

// BinaryExpression like a > 10
type BinaryExpression struct {
	Position Position
	Type     string

	Left     Statement
	Operator Token
	Right    Statement
}

func (l *BinaryExpression) GetPosition() Position {
	return l.Position
}

func (b *BinaryExpression) String() string {
	return fmt.Sprintf("%s %s %s", b.Left.String(), b.Operator.Data, b.Right.String())
}

// SubExpression like a = a && (b * 3), and then '(b * 3)' is SubExpression
type SubExpression struct {
	Position Position
	Type     string

	Expression Statement
}

func (l *SubExpression) GetPosition() Position {
	return l.Position
}

func (b *SubExpression) String() string {
	return fmt.Sprintf("(%s)", b.Expression.String())
}

// Assign like a = 2
type Assign struct {
	Position Position
	Type     string

	Target Statement
	Value  Statement
}

func (l *Assign) GetPosition() Position {
	return l.Position
}

func (a *Assign) String() string {
	switch a.Value.(type) {
	case *Block:
		return fmt.Sprintf("%s = {%s}", a.Target.String(), intent(a.Value.String()))
	case *List:
		return fmt.Sprintf("%s = [%s]", a.Target.String(), intent(a.Value.String()))
	}
	return fmt.Sprintf("%s = %s", a.Target.String(), a.Value.String())
}

// List like [1, 2, 3]
type List struct {
	Position Position
	Type     string

	Values []Statement
}

func (l *List) GetPosition() Position {
	return l.Position
}

func (l *List) String() string {
	var s []string
	for _, item := range l.Values {
		switch item.(type) {
		case *Block:
			s = append(s, fmt.Sprintf("\n{%s}\n", intent(item.String())))
		default:
			s = append(s, item.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ", "))
}

// ListAccess like a[0]
type ListAccess struct {
	Position Position
	Type     string

	Index int
	List  Variable
}

func (l *ListAccess) GetPosition() Position {
	return l.Position
}

func (l *ListAccess) String() string {
	return fmt.Sprintf("%s[%d]", l.List.String(), l.Index)
}

// Block contains many statments
type Block struct {
	Statements []Statement

	Position Position
	Type     string
}

// Position of Block
func (b *Block) GetPosition() Position {
	return b.Position
}

func (b *Block) EndPosition() Position {
	// FIXME: end of statement
	if len(b.Statements) > 0 {
		return b.Statements[len(b.Statements)-1].GetPosition()
	}
	return b.Position
}

func (b *Block) String() string {
	var s []string
	for _, item := range b.Statements {
		s = append(s, item.String())
	}
	return strings.Join(s, "")
}

func (b *Block) Format(root bool) string {
	sb := new(strings.Builder)
	flag := 0
	for index, item := range b.Statements {
		if item == nil {
			break
		}
		switch v := item.(type) {
		case *NewLine:
			if flag < 2 {
				if index != 0 && index != len(b.Statements)-1 {
					sb.WriteString(item.String())
					flag++
				}
			}
		case *Block:
			sb.WriteString(v.Format(false))
			flag = 0
		default:
			flag = 0
			sb.WriteString(item.String())
		}

	}
	if root {
		return sb.String()
	} else {
		return fmt.Sprintf("{\n%s\n}", intent(sb.String()))
	}
}

// Function like test(a, b){}
type Function struct {
	Position Position
	Type     string

	Name       string
	Parameters []Statement
	Body       *Block
}

func (l *Function) GetPosition() Position {
	return l.Position
}

func (f *Function) String() string {
	var args []string
	for _, item := range f.Parameters {
		args = append(args, item.String())
	}
	s := f.Body.Format(false)
	return fmt.Sprintf("%s(%s) %s", f.Name, strings.Join(args, ", "), s)
}

func (f *Function) SignatureString() string {
	var args []string
	for _, item := range f.Parameters {
		args = append(args, item.String())
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", "))
}

// FunctionCall like test(a, b)
type FunctionCall struct {
	Position Position
	Type     string

	Name       string
	Parameters []Statement
}

func (l *FunctionCall) GetPosition() Position {
	return l.Position
}

func (c *FunctionCall) String() string {
	var args []string
	for _, item := range c.Parameters {
		if v, ok := item.(*Block); ok {
			args = append(args, v.Format(false))
		} else {
			args = append(args, item.String())
		}
	}
	return fmt.Sprintf("%s(%s)", c.Name, strings.Join(args, ", "))
}

// ImportFunctionCall like test(a, b)
type ImportFunctionCall struct {
	Position Position
	Type     string

	ModulePath string
	Block      *Block
}

func (l *ImportFunctionCall) GetPosition() Position {
	return l.Position
}

func (c *ImportFunctionCall) String() string {
	return fmt.Sprintf("import(%s)", c.ModulePath)
}

func block(b *Block) string {
	s := collectBlock(b)
	var ss []string
	for _, item := range s {
		ss = append(ss, intent(item))
	}
	return strings.Join(ss, "")
}

// Program means the whole application
type Program struct {
	Statements *Block
}

func (p *Program) String() string {
	return p.Statements.String()
}

// IFStatement like if
type IFStatement struct {
	Position Position
	Type     string

	Condition Statement
	Body      *Block
	Else      *Block
	ElseIf    Statement
}

func (l *IFStatement) GetPosition() Position {
	return l.Position
}

func (i *IFStatement) String() string {
	if i.ElseIf != nil {
		if i.Else != nil && len(i.Else.Statements) != 0 {
			return fmt.Sprintf("if %s {%s} else %s else {%s}", i.Condition.String(), block(i.Body), i.ElseIf.String(), block(i.Else))
		} else {
			return fmt.Sprintf("if %s {%s} else %s", i.Condition.String(), block(i.Body), i.ElseIf.String())
		}
	} else {
		if i.Else != nil && len(i.Else.Statements) != 0 {
			return fmt.Sprintf("if %s {%s} else {%s}", i.Condition.String(), block(i.Body), block(i.Else))
		} else {
			return fmt.Sprintf("if %s {%s}", i.Condition.String(), block(i.Body))
		}
	}
}

// FORStatement like for
type FORStatement struct {
	Position Position
	Type     string

	Iterable IterableExpression
	Block    Block

	CurrentIndex Variable
	CurrentItem  Statement
}

func (l *FORStatement) GetPosition() Position {
	return l.Position
}

func (f *FORStatement) String() string {
	return fmt.Sprintf("for %s, %s in %s {\n%s\n}",
		f.CurrentIndex.String(),
		f.CurrentItem.String(),
		f.Iterable.Name.String(),
		intent(f.Block.String()))
}

// IterableExpression like for in
type IterableExpression struct {
	Position Position
	Type     string

	Name  Variable
	Index int
	Items []Statement
}

func (l *IterableExpression) GetPosition() Position {
	return l.Position
}

func (i *IterableExpression) String() string {
	return ""
}

// Next part of IterableExpression
func (i *IterableExpression) Next() (int, Statement) {
	if i.Index+1 >= len(i.Items) {
		return -1, nil
	}
	item := i.Items[i.Index]
	i.Index++
	return i.Index, item
}

// Break like break in for
type Break struct {
	Position Position
	Type     string
}

func (l *Break) GetPosition() Position {
	return l.Position
}

func (b *Break) String() string {
	return "break"
}

// Continue like continue in for
type Continue struct {
	Position Position
	Type     string
}

func (l *Continue) GetPosition() Position {
	return l.Position
}

func (b *Continue) String() string {
	return "continue"
}

// Return like return varA
type Return struct {
	Position Position
	Type     string

	Value Statement
}

func (l *Return) GetPosition() Position {
	return l.Position
}

func (r *Return) String() string {
	switch v := r.Value.(type) {
	case *Block:
		return fmt.Sprintf("return %s", intent(v.Format(false)))
	}
	return fmt.Sprintf("return %s", r.Value.String())
}

// Field like obj.age
type Field struct {
	Position Position
	Type     string

	Variable Variable
	Value    Statement
}

func (l *Field) GetPosition() Position {
	return l.Position
}

func (f *Field) String() string {
	if v, ok := f.Value.(*Variable); ok && strings.Contains(v.Name, "-") {
		return fmt.Sprintf("%s[%s]", f.Variable.String(), f.Value.String())
	}
	return fmt.Sprintf("%s.%s", f.Variable.String(), f.Value.String())
}

// Boolen like true, false
type Boolen struct {
	Position Position
	Type     string

	Value bool
}

func (l *Boolen) GetPosition() Position {
	return l.Position
}

func (b *Boolen) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// StringExpression like 'hello world !'
type StringExpression struct {
	Position Position
	Type     string

	Value string
}

func (l *StringExpression) GetPosition() Position {
	return l.Position
}

func (s *StringExpression) String() string {
	return s.Value
}

// Comment line for sth
type Comment struct {
	Position Position
	Type     string

	Value string
}

func (l *Comment) GetPosition() Position {
	return l.Position
}

func (c *Comment) String() string {
	return fmt.Sprintf("//%s", c.Value)
}
