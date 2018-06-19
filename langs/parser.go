package langs

import (
	"strconv"
	"fmt"
)

type Parser struct {
	Lexer   *Lexer
	Current Token
}

func NewParser(data []byte) *Parser {
	return &Parser{
		Lexer: NewLexer(data),
	}
}

func (p *Parser) Consume(kind string) Token {
	old := p.Current
	if kind != "" && old.Kind != kind {
		panic(fmt.Sprintf("Invalid token kind: %s  val: %s but except: %s at line: %s, col: %s", old.Kind,
			old.Data, kind, old.Position.Line, old.Position.Col))
	}
	p.Current = p.Lexer.Next()
	return old
}

func (p *Parser) Parse() Block {
	p.Consume("")
	block := Block{}
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

func (p *Parser) ReadStatement() Statement {
	current := p.Consume("")
	switch current.Kind {
	case EOF:
		return nil
	case NAME:
		if current.Data == "return" {

			return &Return{
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
				return &Break{}
			case CONTINUE:
				return &Continue{}
			}
		}
		next := p.Consume("")
		switch next.Kind {
		case EQ:
			return &Assign{
				Target: &Variable{
					Name: current.Data,
				},
				Value: p.ReadExpression(),
			}
		case LParenthese:
			return p.ReadFunction(current.Data)
		case DOT:
			field := &Field{
				Variable: Variable{
					Name: current.Data,
				},
				Value: p.ReadField(),
			}
			if p.Current.Kind == EQ {
				p.Consume(EQ)
				return &Assign{
					Target: field,
					Value:  p.ReadExpression(),
				}
			}
			return field
		}
	default:
		panic(fmt.Sprintf("ReadStatement Unknow Token kind: %s value: %s at line: %d, col: %d", current.Kind,
			current.Data, current.Position.Line, current.Position.Col))
	}
	return nil
}

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

func (p *Parser) ReadFOR() Statement {
	var item FORStatement
	if p.Current.Kind == NAME {
		index := p.Consume(NAME)
		item.CurrentIndex = Variable{
			Name: index.Data,
		}
		p.Consume(COMMA)
		val := p.Consume(NAME)
		item.CurrentItem = &Variable{
			Name: val.Data,
		}
		if p.Current.Data != IN {
			panic("ReadFOR")
		}
		p.Consume(NAME)
		iterable := p.Consume(NAME)
		item.Iterable = IterableExpression{
			Name: Variable{
				Name: iterable.Data,
			},
		}
	} else {
		item.CurrentIndex = Variable{
			Name: "index",
		}
		item.CurrentItem = &Variable{
			Name: "item",
		}
		item.Iterable = IterableExpression{
			Name: Variable{
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
		Name:       fn.Name,
		Parameters: fn.Parameters,
	}
}

func (p *Parser) ReadList() Expresion {
	l := []Expresion{}
	for {
		if p.Current.Kind == RBracket {
			p.Consume(RBracket)
			break
		}
		exp := p.ReadExpression()
		l = append(l, exp)
		if p.Current.Kind == RBracket {
			p.Consume(RBracket)
			break
		}
		p.Consume("")
	}
	return &List{
		Values: l,
	}
}

func (p *Parser) ReadExpression() Expresion {
	current := p.Consume("")
	switch current.Kind {
	case NAME:
		switch p.Current.Kind {
		case PLUS, MINUS, TIMES, DEVIDE:
			return &BinaryExpression{
				Left: &Variable{
					Name: current.Data,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		case LT, LTE, GT, GTE:
			return &BinaryExpression{
				Left: &Variable{
					Name: current.Data,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		case EQ:
			if p.Current.Kind == EQ {
				next := p.Consume(EQ)
				return &BinaryExpression{
					Left: &Variable{
						Name: current.Data,
					},
					Operator: p.Consume(next.Kind),
					Right:    p.ReadExpression(),
				}
			}
		case LParenthese:
			p.Consume(LParenthese)
			fn1 := p.ReadFunction(current.Data)
			switch item := fn1.(type) {
			case *FunctionCall:
				switch p.Current.Kind {
				case MINUS, PLUS, TIMES, DEVIDE:
					return &BinaryExpression{
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
				Variable: Variable{
					Name: current.Data,
				},
				Value: p.ReadField(),
			}
			if p.Current.Kind == EQ {
				p.Consume(EQ)
				return &Assign{
					Target: field,
					Value:  p.ReadExpression(),
				}
			}
			return field
		default:
			if current.Data == "true" {
				return &Boolen{
					Value: true,
				}
			}
			if current.Data == "false" {
				return &Boolen{
					Value: false,
				}
			}
			switch p.Current.Kind {
			case PLUS:
			case MINUS:
				return p.ReadExpression()
			}
			return &Variable{
				Name: current.Data,
			}
		}
	case PLUS:
		return p.ReadExpression()
	case INT:
		value, _ := strconv.Atoi(current.Data)
		switch p.Current.Kind {
		case MINUS, PLUS, TIMES, DEVIDE:
			return &BinaryExpression{
				Left: &Literal{
					Value: value,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		}
		return &Literal{
			Value: value,
		}
	case STRING:
		switch p.Current.Kind {
		case PLUS, MINUS:
			return &BinaryExpression{
				Left: &Literal{
					Value: current.Data,
				},
				Operator: p.Consume(p.Current.Kind),
				Right:    p.ReadExpression(),
			}
		}
		return &Literal{
			Value: current.Data,
		}
	case LParenthese:
		return p.ReadExpression()
	case LBrace:
		return p.ReadDict()
	case LBracket:
		return p.ReadList()
	}
	panic(fmt.Sprintf("Unknow Expression at line: %d, col: %d", current.Position.Line, current.Position.Col))
}

func (p *Parser) ReadDict() Expresion {
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

func (p *Parser) ReadField() Expresion {
	name := p.Consume(NAME)
	if p.Current.Kind == DOT {
		p.Consume(DOT)
		return &Field{
			Variable: Variable{
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
		Name: name.Data,
	}
}
