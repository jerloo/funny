package funny

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// Parser the parser
type Parser struct {
	Lexer   *Lexer
	Current Token

	Tokens []Token

	ContentFile string
}

// NewParser create a new parser
func NewParser(data []byte, file string) *Parser {
	return &Parser{
		Lexer:       NewLexer(data, file),
		ContentFile: file,
	}
}

// Consume get next token
func (p *Parser) Consume(kind string) Token {
	old := p.Current
	p.Tokens = append(p.Tokens, old)
	if kind != "" && old.Kind != kind {
		return old
		// panic(P(fmt.Sprintf("Invalid token kind %s", old.String()), old.Position))
	}
	p.Current = p.Lexer.Next()
	return old
}

// Parse parse to statements
func (p *Parser) Parse() (block *Block, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	block = &Block{
		Type: STBlock,
	}
	p.Consume("")
	for {
		if p.Current.Kind == EOF {
			break
		}
		element := p.ReadStatement()
		if element == nil {
			break
		}
		block.Statements = append(block.Statements, element)
	}
	return
}

// ReadStatement get next statement
func (p *Parser) ReadStatement() Statement {
	current := p.Consume("")
	switch current.Kind {
	case EOF:
		return nil
	case NAME:
		if current.Data == "return" {
			return &Return{
				Position: current.Position,
				Value:    p.ReadExpression(),
				Type:     STReturn,
			}
		}
		kind, ok := Keywords[current.Data]
		if ok {
			switch kind {
			case IF:
				return p.ReadIF()
			case FOR:
				return p.ReadFOR()
			case BREAK:
				return &Break{
					Position: current.Position,
					Type:     STBreak,
				}
			case CONTINUE:
				return &Continue{
					Position: current.Position,
					Type:     STContinue,
				}
			}
		}
		next := p.Consume("")
		switch next.Kind {
		case EQ:
			return &Assign{
				Position: current.Position,
				Target: &Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Value: p.ReadExpression(),
				Type:  STAssign,
			}
		case LParenthese:
			return p.ReadFunction(current.Data)
		case DOT:
			field := &Field{
				Position: current.Position,
				Variable: Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Value: p.ReadField(),
				Type:  STField,
			}
			if p.Current.Kind == EQ {
				p.Consume(EQ)
				return &Assign{
					Position: current.Position,
					Target:   field,
					Value:    p.ReadExpression(),
					Type:     STAssign,
				}
			}
			return field
		case LBracket:
			key := p.Consume(STRING)
			p.Consume(RBracket)
			field := &Field{
				Position: current.Position,
				Variable: Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Value: &Variable{
					Position: current.Position,
					Name:     key.Data,
					Type:     STVariable,
				},
				Type: STField,
			}
			switch p.Current.Kind {
			case EQ:
				p.Consume(EQ)
				return &Assign{
					Position: current.Position,
					Target:   field,
					Value:    p.ReadExpression(),
					Type:     STAssign,
				}
			case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
				return &BinaryExpression{
					Position: current.Position,
					Left:     field,
					Operator: p.Consume(p.Current.Kind),
					Right:    p.ReadExpression(),
					Type:     STBinaryExpression,
				}
			}
		}
	case COMMENT:
		return &Comment{
			Position: current.Position,
			Value:    current.Data,
			Type:     STComment,
		}
	case NEW_LINE:
		return &NewLine{
			Position: current.Position,
			Type:     STNewLine,
		}
	case STRING:
		switch p.Current.Kind {
		case EQ:
			p.Consume(EQ)
			return &Assign{
				Position: current.Position,
				Target: &Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Value: p.ReadExpression(),
				Type:  STAssign,
			}
		}
	default:
		panic(P(fmt.Sprintf("ReadStatement Unknow Token %s", current.String()), current.Position))
	}
	return nil
}

// ReadIF get next if statement
func (p *Parser) ReadIF() Statement {
	item := &IFStatement{
		Position: p.Current.Position,
		Type:     STIfStatement,
	}

	item.Condition = p.ReadExpression()

	// if body
	p.Consume(LBrace)

	for {
		if p.Current.Kind == RBrace {
			p.Consume(RBrace)
			break
		}
		if item.Body == nil {
			item.Body = &Block{
				Position: p.Current.Position,
			}
		}
		item.Body.Statements = append(item.Body.Statements, p.ReadStatement())
	}

	// else body
	if p.Current.Kind == NAME && p.Current.Data == ELSE {
		p.Consume("")
		if p.Current.Kind == NAME && p.Current.Data == IF {
			p.Consume("")
			item.ElseIf = p.ReadIF()
		} else {
			p.Consume(LBrace)
			for {
				if p.Current.Kind == RBrace {
					break
				}
				if item.Else == nil {
					item.Else = &Block{
						Position: p.Current.Position,
					}
				}
				item.Else.Statements = append(item.Else.Statements, p.ReadStatement())
			}
			p.Consume(RBrace)
		}
	}
	return item
}

// ReadFOR read for statement
func (p *Parser) ReadFOR() Statement {
	var item FORStatement
	if p.Current.Kind == NAME {
		index := p.Consume(NAME)
		item.CurrentIndex = Variable{
			Position: p.Current.Position,
			Name:     index.Data,
			Type:     STVariable,
		}
		p.Consume(COMMA)
		val := p.Consume(NAME)
		item.CurrentItem = &Variable{
			Position: p.Current.Position,
			Name:     val.Data,
			Type:     STVariable,
		}
		if p.Current.Data != IN {
			panic(P("for must has in part", p.Current.Position))
		}
		p.Consume(NAME)
		iterable := p.Consume(NAME)
		item.Iterable = IterableExpression{
			Position: p.Current.Position,
			Name: Variable{
				Position: p.Current.Position,
				Name:     iterable.Data,
				Type:     STVariable,
			},
			Type: STIterableExpression,
		}
	} else {
		item.CurrentIndex = Variable{
			Position: p.Current.Position,
			Name:     "index",
			Type:     STVariable,
		}
		item.CurrentItem = &Variable{
			Position: p.Current.Position,
			Name:     "item",
			Type:     STVariable,
		}
		item.Iterable = IterableExpression{
			Position: p.Current.Position,
			Name: Variable{
				Position: p.Current.Position,
				Name:     "items",
				Type:     STVariable,
			},
			Type: STIterableExpression,
		}
	}
	p.Consume(LBrace)
	for {
		if p.Current.Kind == RBrace {
			p.Consume(RBrace)
			break
		}
		sub := p.ReadStatement()
		item.Block.Statements = append(item.Block.Statements, sub)
	}

	return &item
}

// ReadFunctionCall read function statement
func (p *Parser) ReadFunctionCall(name string) Statement {
	pos := p.Current.Position
	fn := &Function{
		Body: &Block{
			Type: STBlock,
		},
		Type: STFunction,
	}
	fn.Name = name
	for {
		if p.Current.Kind == COMMA {
			p.Consume(COMMA)
			continue
		}
		if p.Current.Kind == RParenthese {
			p.Consume(RParenthese)
			break
		}
		fn.Parameters = append(fn.Parameters, p.ReadExpression())
	}

	if fn.Name == "import" {
		if len(fn.Parameters) == 0 {
			panic(P("import module path can not be empty", fn.Position))
		}
		arg := fn.Parameters[0]
		moduleArg, ok := arg.(*Literal)
		if !ok {
			panic(P(fmt.Sprintf("import module path not string type %s", fn.Parameters[0].String()), p.Current.Position))
		}
		moduleFileName := moduleArg.Value.(string)
		if strings.HasPrefix(moduleFileName, ".") {
			if p.ContentFile == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					panic(P(fmt.Sprintf("import module path not found %s", moduleFileName), p.Current.Position))
				}
				moduleFileName = path.Join(currentDir, moduleFileName)
			} else {
				currentDir := path.Dir(p.ContentFile)
				moduleFileName = path.Join(currentDir, moduleFileName)
			}
		} else {
			panic(P(fmt.Sprintf("import module path not found %s", fn.Parameters[0].String()), p.Current.Position))
		}
		importData, err := os.ReadFile(moduleFileName)
		if err != nil {
			panic(P(fmt.Sprintf("import module path not found %s", fn.Parameters[0].String()), p.Current.Position))
		}
		importParser := NewParser(importData, moduleFileName)
		block, err := importParser.Parse()
		if err != nil {
			panic(err)
		}
		return &ImportFunctionCall{
			Position:   p.Current.Position,
			ModulePath: fn.Parameters[0].String(),
			Block:      block,
			Type:       STFunctionCall,
		}
	}
	return &FunctionCall{
		Position:   pos,
		Name:       fn.Name,
		Parameters: fn.Parameters,
		Type:       STFunctionCall,
	}
}

// ReadFunction read function statement
func (p *Parser) ReadFunction(name string) Statement {
	pos := p.Current.Position
	fn := &Function{
		Body: &Block{
			Type: STBlock,
		},
		Type: STFunction,
	}
	fn.Name = name
	for {
		if p.Current.Kind == COMMA {
			p.Consume(COMMA)
			continue
		}
		if p.Current.Kind == RParenthese {
			p.Consume(RParenthese)
			break
		}
		fn.Parameters = append(fn.Parameters, p.ReadExpression())
	}
	if p.Current.Kind == LBrace {
		p.Consume(LBrace)
		for {
			if p.Current.Kind == RBrace {
				p.Consume(RBrace)
				break
			}
			sub := p.ReadStatement()
			if sub == nil {
				break
			}
			fn.Body.Statements = append(fn.Body.Statements, sub)
		}
		return fn
	}
	if fn.Name == "import" {
		if len(fn.Parameters) == 0 {
			panic(P("import module path can not be empty", fn.Position))
		}
		arg := fn.Parameters[0]
		moduleArg, ok := arg.(*Literal)
		if !ok {
			panic(P(fmt.Sprintf("import module path not string type %s", fn.Parameters[0].String()), p.Current.Position))
		}
		moduleFileName := moduleArg.Value.(string)
		if strings.HasPrefix(moduleFileName, ".") {
			if p.ContentFile == "" {
				currentDir, err := os.Getwd()
				if err != nil {
					panic(P(fmt.Sprintf("import module path not found %s", moduleFileName), p.Current.Position))
				}
				moduleFileName = path.Join(currentDir, moduleFileName)
			} else {
				currentDir := path.Dir(p.ContentFile)
				moduleFileName = path.Join(currentDir, moduleFileName)
			}
		} else {
			panic(P(fmt.Sprintf("import module path not found %s", fn.Parameters[0].String()), p.Current.Position))
		}
		importData, err := os.ReadFile(moduleFileName)
		if err != nil {
			panic(P(fmt.Sprintf("import module path not found %s", fn.Parameters[0].String()), p.Current.Position))
		}
		importParser := NewParser(importData, moduleFileName)
		block, err := importParser.Parse()
		if err != nil {
			panic(err)
		}
		return &ImportFunctionCall{
			Position:   p.Current.Position,
			ModulePath: fn.Parameters[0].String(),
			Block:      block,
			Type:       STFunctionCall,
		}
	}
	return &FunctionCall{
		Position:   pos,
		Name:       fn.Name,
		Parameters: fn.Parameters,
		Type:       STFunctionCall,
	}
}

// ReadList read list expression
func (p *Parser) ReadList() Statement {
	startPosition := p.Current.Position
	l := []Statement{}
	for {
		if p.Current.Kind == NEW_LINE {
			p.Consume(NEW_LINE)
			continue
		} else if p.Current.Kind == LBrace {
			p.Consume(LBrace)
			dic := p.ReadDict()
			l = append(l, dic)
			continue
		} else if p.Current.Kind == COMMA {
			p.Consume(COMMA)
			continue
		} else if p.Current.Kind == RBracket {
			p.Consume(RBracket)
			break
		}
		exp := p.ReadExpression()
		l = append(l, exp)
		// p.Consume("")
	}

	return &List{
		Position: startPosition,
		Values:   l,
		Type:     STList,
	}
}

// ReadExpression read next expression
func (p *Parser) ReadExpression() Statement {
	current := p.Consume("")
	switch current.Kind {
	case NAME:
		if current.Data == IN {
			return &BinaryExpression{
				Position: current.Position,
				Left: &Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Operator: current,
				Right:    p.ReadExpression(),
				Type:     STBinaryExpression,
			}
		}
		switch p.Current.Kind {
		case PLUS, MINUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ, NAME:
			return &BinaryExpression{
				Position: current.Position,
				Left: &Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
				Type:     STBinaryExpression,
			}
		case LParenthese:
			p.Consume(LParenthese)
			fn1 := p.ReadFunctionCall(current.Data)
			switch item := fn1.(type) {
			case *FunctionCall:
				switch p.Current.Kind {
				case MINUS, PLUS, TIMES, DEVIDE:
					return &BinaryExpression{
						Position: current.Position,
						Left:     item,
						Operator: p.Consume(p.Current.Kind),
						Right:    p.ReadExpression(),
						Type:     STBinaryExpression,
					}
				}
			}
			return fn1
		case DOT:
			p.Consume(DOT)
			field := &Field{
				Position: current.Position,
				Variable: Variable{
					Position: current.Position,
					Name:     current.Data,
					Type:     STVariable,
				},
				Value: p.ReadField(),
				Type:  STField,
			}
			switch p.Current.Kind {
			case EQ:
				p.Consume(EQ)
				return &Assign{
					Position: current.Position,
					Target:   field,
					Value:    p.ReadExpression(),
					Type:     STAssign,
				}
			case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
				return &BinaryExpression{
					Position: current.Position,
					Left:     field,
					Operator: p.Consume(p.Current.Kind),
					Right:    p.ReadExpression(),
					Type:     STBinaryExpression,
				}
			}
			return field
		case LBracket:
			p.Consume(LBracket)
			var exp Statement
			if p.Current.Kind == NAME {
				// Field access
				key := p.Consume("")
				p.Consume(RBracket)
				exp = &Field{
					Position: current.Position,
					Variable: Variable{
						Position: current.Position,
						Name:     current.Data,
						Type:     STVariable,
					},
					Value: &Variable{
						Name:     key.Data,
						Type:     STVariable,
						Position: key.Position,
					},
					Type: STField,
				}
			} else if p.Current.Kind == STRING {
				// Field access
				key := p.Consume("")
				p.Consume(RBracket)
				exp = &Field{
					Position: current.Position,
					Variable: Variable{
						Position: current.Position,
						Name:     current.Data,
						Type:     STVariable,
					},
					Value: &StringExpression{
						Type:     STStringExpression,
						Value:    key.Data,
						Position: key.Position,
					},
					Type: STField,
				}
			} else if p.Current.Kind == INT {
				token := p.Consume(INT)
				indexStr := token.Data
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					panic(P("Bad list index ", token.Position))
				}
				exp = &ListAccess{
					Position: current.Position,
					List: Variable{
						Position: current.Position,
						Name:     current.Data,
						Type:     STVariable,
					},
					Index: index,
					Type:  STListAccess,
				}
				p.Consume(RBracket)
			} else {
				panic(P(fmt.Sprintf("Unknow Kind Reading Field %s", p.Current.Kind), p.Current.Position))
			}

			switch p.Current.Kind {
			case EQ:
				p.Consume(EQ)
				return &Assign{
					Position: current.Position,
					Target:   exp,
					Value:    p.ReadExpression(),
					Type:     STAssign,
				}
			case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
				return &BinaryExpression{
					Position: current.Position,
					Left:     exp,
					Operator: p.Consume(p.Current.Kind),
					Right:    p.ReadExpression(),
					Type:     STBinaryExpression,
				}
			default:
				return exp
			}

		default:
			if current.Data == "true" {
				return &Boolen{
					Position: current.Position,
					Value:    true,
					Type:     STBoolean,
				}
			}
			if current.Data == "false" {
				return &Boolen{
					Position: current.Position,
					Value:    false,
					Type:     STBoolean,
				}
			}
			switch p.Current.Kind {
			case PLUS:
			case MINUS:
				return p.ReadExpression()
			}
			return &Variable{
				Position: current.Position,
				Name:     current.Data,
				Type:     STVariable,
			}
		}
	case PLUS:
		return p.ReadExpression()
	case INT:
		value, _ := strconv.Atoi(current.Data)
		switch p.Current.Kind {
		case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ, NAME:
			return &BinaryExpression{
				Position: current.Position,
				Left: &Literal{
					Position: current.Position,
					Value:    value,
					Type:     STLiteral,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
				Type:     STBinaryExpression,
			}
		}
		return &Literal{
			Position: current.Position,
			Value:    value,
			Type:     STLiteral,
		}
	case STRING:
		switch p.Current.Kind {
		case PLUS, MINUS:
			return &BinaryExpression{
				Position: current.Position,
				Left: &Literal{
					Position: current.Position,
					Value:    current.Data,
					Type:     STLiteral,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
				Type:     STBinaryExpression,
			}
		}
		return &Literal{
			Position: current.Position,
			Value:    current.Data,
			Type:     STLiteral,
		}
	case LParenthese:
		p.Consume(LParenthese)
		exp := &SubExpression{
			Position:   p.Current.Position,
			Type:       STSubExpression,
			Expression: p.ReadExpression(),
		}
		p.Consume(RParenthese)
		switch p.Current.Kind {
		case MINUS, PLUS, TIMES, DEVIDE:
			return &BinaryExpression{
				Position: p.Current.Position,
				Type:     STBinaryExpression,
				Left:     exp,
				Operator: p.Consume(""),
				Right:    p.ReadExpression(),
			}
		}
		return exp
	case LBrace:
		return p.ReadDict()
	case LBracket:
		return p.ReadList()
	}
	panic(P(fmt.Sprintf("Unknow Expression Data: %s", current.Data), current.Position))
}

// ReadDict read dict expression
func (p *Parser) ReadDict() Statement {
	b := &Block{
		Type: STBlock,
	}
	for {
		if p.Current.Kind == RBrace {
			p.Consume(RBrace)
			break
		}
		sub := p.ReadStatement()
		b.Statements = append(b.Statements, sub)
	}
	return b
}

// ReadField read field expression
func (p *Parser) ReadField() Statement {
	name := p.Consume(NAME)
	if p.Current.Kind == DOT {
		p.Consume(DOT)
		return &Field{
			Position: p.Current.Position,
			Variable: Variable{
				Position: p.Current.Position,
				Name:     name.Data,
				Type:     STVariable,
			},
			Value: p.ReadField(),
			Type:  STField,
		}
	}
	if p.Current.Kind == LParenthese {
		p.Consume(LParenthese)
		return p.ReadFunction(name.Data)
	}
	return &Variable{
		Position: p.Current.Position,
		Name:     name.Data,
		Type:     STVariable,
	}
}
