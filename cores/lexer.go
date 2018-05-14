package cores

import (
	"unicode/utf8"
	"fmt"
)

type Position struct {
	Line int
	Col  int
}

func (p *Position) String() string {
	return fmt.Sprintf("[Position] Line: %4d, Col: %4d", p.Line, p.Col)
}

type Token struct {
	Position Position
	Kind     string
	Data     string
}

func (t *Token) String() string {
	return fmt.Sprintf("[Token] Kind: %6s, %6s, Data: %6s", t.Kind, t.Position.String(), t.Data)
}

type Lexer struct {
	Offset     int
	CurrentPos Position

	SaveOffset int
	SavePos    Position
	Data       []byte
	Elements   []Token
}

func NewLexer(data []byte) *Lexer {
	return &Lexer{
		Data: data,
		CurrentPos: Position{
			Line: 1,
			Col:  1,
		},
	}
}

func (l *Lexer) LA(n int) rune {
	offset := l.Offset
	for n >= 0 {
		ch, size := utf8.DecodeRune(l.Data[offset:])
		if offset+size > len(l.Data) {
			return -1
		}
		offset += size
		if n == 0 {
			return ch
		}
		n--
	}
	return -1
}

func (l *Lexer) Consume(n int) rune {
	for n > 0 {
		ch, size := utf8.DecodeRune(l.Data[l.Offset:])
		if l.Offset+size > len(l.Data) {
			return -1
		}
		l.Offset += size
		if n == 0 {
			return ch
		}
		n--
	}
	return -1
}

func (l *Lexer) CreateToken(kind string) Token {
	st := l.Data[l.SaveOffset+1 : l.Offset+1]
	token := Token{
		Kind:     kind,
		Data:     string(st),
		Position: l.CurrentPos,
	}
	l.CurrentPos.Col += len(token.Data)
	return token
}

func (l *Lexer) NewLine() {
	l.CurrentPos.Col = 1
	l.CurrentPos.Line++
}

func (l *Lexer) ReadWhiteAndComments() {
DONE:
	for {
		ch := l.LA(1)
		switch ch {
		case '\n':
			l.Consume(1)
			l.NewLine()
			break DONE
		case ' ', '\t':
			l.Consume(1)
		default:
			break DONE
		}
	}
}

func isNameStart(ch rune) bool {
	chString := string(ch)
	fmt.Sprintf(chString)
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func (l *Lexer) ReadInt() Token {
	for {
		ch := l.LA(1)
		switch ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			l.Consume(1)
		default:
			return l.CreateToken(INT)
			break
		}
	}
	return l.CreateToken(EOF)
}

func (l *Lexer) Reset() {
	l.SaveOffset = l.Offset
	l.SavePos = l.CurrentPos
}

func (l *Lexer) Next() Token {
	for {
		l.Reset()
		ch := l.LA(1)
		chString := string(ch)
		fmt.Sprintf(chString)
		switch ch {
		case -1:
			l.Consume(1)
			return l.CreateToken(EOF)
		case '\n', ' ', '\t':
			l.Consume(1)
			l.ReadWhiteAndComments()
		case '/':
			if chNext := l.LA(2); chNext == '/' {
				l.Consume(2)
				l.ReadWhiteAndComments()
			}
			return l.CreateToken(DEVIDE)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return l.ReadInt()
		case '=':
			l.Consume(1)
			return l.CreateToken(EQ)
		case '+':
			l.Consume(1)
			return l.CreateToken(PLUS)
		case '-':
			l.Consume(1)
			return l.CreateToken(MINUS)
		case '*':
			l.Consume(1)
			return l.CreateToken(TIMES)
		case '(':
			l.Consume(1)
			return l.CreateToken(LParenthese)
		case ')':
			l.Consume(1)
			return l.CreateToken(RParenthese)
		case '[':
			l.Consume(1)
			return l.CreateToken(LBracket)
		case ']':
			l.Consume(1)
			return l.CreateToken(RBracket)
		case '{':
			l.Consume(1)
			return l.CreateToken(LBrace)
		case '}':
			l.Consume(1)
			return l.CreateToken(RBrace)
		case ',':
			l.Consume(1)
			return l.CreateToken(COMMA)
		case '.':
			l.Consume(1)
			return l.CreateToken(DOT)
		case '>':
			if l.LA(2) == '=' {
				l.Consume(2)
				return l.CreateToken(GTE)
			}
			l.Consume(1)
			return l.CreateToken(GT)
		case '<':
			if l.LA(2) == '=' {
				l.Consume(2)
				return l.CreateToken(LTE)
			}
			l.Consume(1)
			return l.CreateToken(LT)
		case '!':
			if l.LA(2) == '=' {
				l.Consume(2)
				return l.CreateToken(NOTEQ)
			}
		default:

			if isNameStart(ch) {
				for {
					chNext := l.LA(1)
					chNS := string(chNext)
					fmt.Sprintf("%s", chNS)
					if !isNameStart(chNext) {
						return l.CreateToken(NAME)
					}
					l.Consume(1)
				}
			}
			l.Consume(1)
			return l.CreateToken(EOF)
		}
	}
}
