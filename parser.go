package funny

import (
	"fmt"
	"strconv"
)

// Parser the parser
type Parser struct {
	Lexer   *Lexer
	Current Token

	Tokens []Token
}

// NewParser create a new parser
func NewParser(data []byte) *Parser {
	return &Parser{
		Lexer: NewLexer(data),
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
func (p *Parser) Parse() Block {
	block := Block{}
	p.Consume("")
	for {
		if p.Current.Kind == EOF {
			break
		}
		element := p.ReadStatement()
		if element == nil {
			break
		}
		block = append(block, element)
	}
	return block
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
				pos:   current.Position,
				Value: p.ReadExpression(),
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
					pos: current.Position,
				}
			case CONTINUE:
				return &Continue{
					pos: current.Position,
				}
			}
		}
		next := p.Consume("")
		switch next.Kind {
		case EQ:
			return &Assign{
				pos: current.Position,
				Target: &Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Value: p.ReadExpression(),
			}
		case LParenthese:
			return p.ReadFunction(current.Data)
		case DOT:
			field := &Field{
				pos: current.Position,
				Variable: Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Value: p.ReadField(),
			}
			if p.Current.Kind == EQ {
				p.Consume(EQ)
				return &Assign{
					pos:    current.Position,
					Target: field,
					Value:  p.ReadExpression(),
				}
			}
			return field
		case LBracket:
			key := p.Consume(STRING)
			p.Consume(RBracket)
			field := &Field{
				pos: current.Position,
				Variable: Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Value: &Variable{
					pos:  current.Position,
					Name: key.Data,
				},
			}
			switch p.Current.Kind {
			case EQ:
				p.Consume(EQ)
				return &Assign{
					pos:    current.Position,
					Target: field,
					Value:  p.ReadExpression(),
				}
			case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
				return &BinaryExpression{
					pos:      current.Position,
					Left:     field,
					Operator: p.Consume(p.Current.Kind),
					Right:    p.ReadExpression(),
				}
			}
		}
	case COMMENT:
		return &Comment{
			pos:   current.Position,
			Value: current.Data,
		}
	case NEW_LINE:
		return &NewLine{
			pos: current.Position,
		}
	case STRING:
		switch p.Current.Kind {
		case EQ:
			p.Consume(EQ)
			return &Assign{
				pos: current.Position,
				Target: &Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Value: p.ReadExpression(),
			}
		}
	default:
		panic(P(fmt.Sprintf("ReadStatement Unknow Token %s", current.String()), current.Position))
	}
	return nil
}

// ReadIF get next if statement
func (p *Parser) ReadIF() Statement {
	var item IFStatement

	item.Condition = p.ReadExpression()

	// if body
	p.Consume(LBrace)

	for {
		if p.Current.Kind == RBrace {
			p.Consume(RBrace)
			break
		}
		item.Body = append(item.Body, p.ReadStatement())
	}

	// else body
	if p.Current.Kind == NAME && p.Current.Data == ELSE {
		p.Consume("")
		p.Consume(LBrace)
		for {
			if p.Current.Kind == RBrace {
				break
			}
			item.Else = append(item.Else, p.ReadStatement())
		}
		p.Consume(RBrace)
	}
	return &item
}

// ReadFOR read for statement
func (p *Parser) ReadFOR() Statement {
	var item FORStatement
	if p.Current.Kind == NAME {
		index := p.Consume(NAME)
		item.CurrentIndex = Variable{
			pos:  p.Current.Position,
			Name: index.Data,
		}
		p.Consume(COMMA)
		val := p.Consume(NAME)
		item.CurrentItem = &Variable{
			pos:  p.Current.Position,
			Name: val.Data,
		}
		if p.Current.Data != IN {
			panic(P("for must has in part", p.Current.Position))
		}
		p.Consume(NAME)
		iterable := p.Consume(NAME)
		item.Iterable = IterableExpression{
			pos: p.Current.Position,
			Name: Variable{
				pos:  p.Current.Position,
				Name: iterable.Data,
			},
		}
	} else {
		item.CurrentIndex = Variable{
			pos:  p.Current.Position,
			Name: "index",
		}
		item.CurrentItem = &Variable{
			pos:  p.Current.Position,
			Name: "item",
		}
		item.Iterable = IterableExpression{
			pos: p.Current.Position,
			Name: Variable{
				pos:  p.Current.Position,
				Name: "items",
			},
		}
	}
	p.Consume(LBrace)
	for {
		if p.Current.Kind == RBrace {
			p.Consume(RBrace)
			break
		}
		sub := p.ReadStatement()
		item.Block = append(item.Block, sub)
	}

	return &item
}

// ReadFunction read function statement
func (p *Parser) ReadFunction(name string) Statement {
	var fn Function
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
			fn.Body = append(fn.Body, sub)
		}
		return &fn

	}
	return &FunctionCall{
		pos:        p.Current.Position,
		Name:       fn.Name,
		Parameters: fn.Parameters,
	}
}

// ReadList read list expression
func (p *Parser) ReadList() Expression {
	l := []Expression{}
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
		pos:    p.Current.Position,
		Values: l,
	}
}

// ReadExpression read next expression
func (p *Parser) ReadExpression() Expression {
	current := p.Consume("")
	switch current.Kind {
	case NAME:
		switch p.Current.Kind {
		case PLUS, MINUS, TIMES, DEVIDE:
			return &BinaryExpression{
				pos: current.Position,
				Left: &Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		case LT, LTE, GT, GTE, DOUBLE_EQ:
			return &BinaryExpression{
				pos: current.Position,
				Left: &Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		case LParenthese:
			p.Consume(LParenthese)
			fn1 := p.ReadFunction(current.Data)
			switch item := fn1.(type) {
			case *FunctionCall:
				switch p.Current.Kind {
				case MINUS, PLUS, TIMES, DEVIDE:
					return &BinaryExpression{
						pos:      current.Position,
						Left:     item,
						Operator: p.Consume(p.Current.Kind),
						Right:    p.ReadExpression(),
					}
				}
			}
			return fn1
		case DOT:
			p.Consume(DOT)
			field := &Field{
				pos: current.Position,
				Variable: Variable{
					pos:  current.Position,
					Name: current.Data,
				},
				Value: p.ReadField(),
			}
			switch p.Current.Kind {
			case EQ:
				p.Consume(EQ)
				return &Assign{
					pos:    current.Position,
					Target: field,
					Value:  p.ReadExpression(),
				}
			case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
				return &BinaryExpression{
					pos:      current.Position,
					Left:     field,
					Operator: p.Consume(p.Current.Kind),
					Right:    p.ReadExpression(),
				}
			}
			return field
		case LBracket:
			p.Consume(LBracket)
			var exp Expression
			if p.Current.Kind == STRING {
				// Field access
				key := p.Consume(STRING)
				p.Consume(RBracket)
				exp = &Field{
					pos: current.Position,
					Variable: Variable{
						pos:  current.Position,
						Name: current.Data,
					},
					Value: &Variable{
						Name: key.Data,
					},
				}
			} else if p.Current.Kind == INT {
				indexStr := p.Consume(INT).Data
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					panic("Bad list index ")
				}
				exp = &ListAccess{
					pos: current.Position,
					List: Variable{
						pos:  current.Position,
						Name: current.Data,
					},
					Index: index,
				}
				p.Consume(RBracket)
			} else {
				panic(P(fmt.Sprintf("Unknow Kind %s", p.Current.Kind), p.Current.Position))
			}

			switch p.Current.Kind {
			case EQ:
				p.Consume(EQ)
				return &Assign{
					pos:    current.Position,
					Target: exp,
					Value:  p.ReadExpression(),
				}
			case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
				return &BinaryExpression{
					pos:      current.Position,
					Left:     exp,
					Operator: p.Consume(p.Current.Kind),
					Right:    p.ReadExpression(),
				}
			default:
				return exp
			}

		default:
			if current.Data == "true" {
				return &Boolen{
					pos:   current.Position,
					Value: true,
				}
			}
			if current.Data == "false" {
				return &Boolen{
					pos:   current.Position,
					Value: false,
				}
			}
			switch p.Current.Kind {
			case PLUS:
			case MINUS:
				return p.ReadExpression()
			}
			return &Variable{
				pos:  current.Position,
				Name: current.Data,
			}
		}
	case PLUS:
		return p.ReadExpression()
	case INT:
		value, _ := strconv.Atoi(current.Data)
		switch p.Current.Kind {
		case MINUS, PLUS, TIMES, DEVIDE, LT, LTE, GT, GTE, DOUBLE_EQ:
			return &BinaryExpression{
				pos: current.Position,
				Left: &Literal{
					pos:   current.Position,
					Value: value,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		}
		return &Literal{
			pos:   current.Position,
			Value: value,
		}
	case STRING:
		switch p.Current.Kind {
		case PLUS, MINUS:
			return &BinaryExpression{
				pos: current.Position,
				Left: &Literal{
					pos:   current.Position,
					Value: current.Data,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		}
		return &Literal{
			pos:   current.Position,
			Value: current.Data,
		}
	case LParenthese:
		return p.ReadExpression()
	case LBrace:
		return p.ReadDict()
	case LBracket:
		return p.ReadList()
	}
	panic(P(fmt.Sprintf("Unknow Expression Data: %s", current.Data), current.Position))
}

// ReadDict read dict expression
func (p *Parser) ReadDict() Expression {
	var b Block
	for {
		if p.Current.Kind == RBrace {
			p.Consume(RBrace)
			break
		}
		sub := p.ReadStatement()
		b = append(b, sub)
	}
	return &b
}

// ReadField read field expression
func (p *Parser) ReadField() Expression {
	name := p.Consume(NAME)
	if p.Current.Kind == DOT {
		p.Consume(DOT)
		return &Field{
			pos: p.Current.Position,
			Variable: Variable{
				pos:  p.Current.Position,
				Name: name.Data,
			},
			Value: p.ReadField(),
		}
	}
	if p.Current.Kind == LParenthese {
		p.Consume(LParenthese)
		return p.ReadFunction(name.Data)
	}
	return &Variable{
		pos:  p.Current.Position,
		Name: name.Data,
	}
}
