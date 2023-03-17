// Package lexer tokenizes input and lets follow up phases in compiler,
// such as parser, iterate over tokens via Lexer.Next() function. The
// lexer package also exposes a Run method for debugging the lexing
// phase only.
package lexer

import (
	"strconv"
	"strings"
	"unicode"
)

func Run(input string) string {
	l := New(input)
	tok := l.Next()
	var sb strings.Builder
	for ; tok.Type != EOF; tok = l.Next() {
		sb.WriteString(tok.String() + "\n")
	}
	sb.WriteString(tok.String() + "\n")
	return sb.String()
}

type Lexer struct {
	input []rune
	cur   rune // current rune under examination
	pos   int  // current position in input (points to current rune)
	line  int
	col   int
}

func New(input string) *Lexer {
	return &Lexer{input: []rune(input), pos: -1, line: 1}
}

func (l *Lexer) Next() *Token {
	l.advance()

	tok := &Token{
		Offset: l.pos,
		Line:   l.line,
		Col:    l.col,
	}
	switch l.cur {
	case ' ', '\t':
		l.consumeHorizontalWhitespace()
		return tok.SetType(WS)
	case '=':
		if l.peekRune() == '=' {
			l.advance()
			return tok.SetType(EQ)
		}
		return tok.SetType(ASSIGN)
	case '+':
		return tok.SetType(PLUS)
	case '-':
		return tok.SetType(MINUS)
	case '!':
		if l.peekRune() == '=' {
			l.advance()
			return tok.SetType(NOT_EQ)
		}
		return tok.SetType(BANG)
	case '/':
		if l.peekRune() == '/' {
			return tok.SetType(COMMENT).SetLiteral(l.readComment())
		}
		return tok.SetType(SLASH)
	case '*':
		return tok.SetType(ASTERISK)
	case '%':
		return tok.SetType(PERCENT)
	case '<':
		if l.peekRune() == '=' {
			l.advance()
			return tok.SetType(LTEQ)
		}
		return tok.SetType(LT)
	case '>':
		if l.peekRune() == '=' {
			l.advance()
			return tok.SetType(GTEQ)
		}
		return tok.SetType(GT)
	case ':':
		if l.peekRune() == '=' {
			l.advance()
			return tok.SetType(DECLARE)
		}
		return tok.SetType(COLON)
	case '{':
		return tok.SetType(LCURLY)
	case '}':
		return tok.SetType(RCURLY)
	case '(':
		return tok.SetType(LPAREN)
	case ')':
		return tok.SetType(RPAREN)
	case '[':
		return tok.SetType(LBRACKET)
	case ']':
		return tok.SetType(RBRACKET)
	case '\n':
		return tok.SetType(NL)
	case '.':
		if l.peekRune() == '.' && l.peekRune2() == '.' {
			l.advance()
			l.advance()
			return tok.SetType(DOT3)
		}
		return tok.SetType(DOT)
	case '"':
		literal, err := l.readString()
		// strconv.Unquote error
		if err != nil {
			return tok.SetType(ILLEGAL).SetLiteral("invalid string")
		}
		return tok.SetType(STRING_LIT).SetLiteral(literal)
	case 0:
		return tok.SetType(EOF)
	}
	if isLetter(l.cur) {
		literal := l.readIdent()
		tokenType := LookupKeyword(literal)
		if tokenType == IDENT {
			return tok.SetType(IDENT).SetLiteral(literal)
		}
		return tok.SetType(tokenType)
	}
	if isDigit(l.cur) {
		return tok.SetType(NUM_LIT).SetLiteral(l.readNum())
	}

	return tok.SetType(ILLEGAL).SetLiteral(string(l.cur))
}

func (l *Lexer) advance() {
	if l.cur == '\n' {
		l.line++
		l.col = 0
	}
	l.pos++
	l.col++
	l.cur = l.lookAt(l.pos)
}

func (l *Lexer) peekRune() rune {
	return l.lookAt(l.pos + 1)
}

func (l *Lexer) peekRune2() rune {
	return l.lookAt(l.pos + 2)
}

func (l *Lexer) lookAt(pos int) rune {
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func (l *Lexer) consumeHorizontalWhitespace() {
	for pr := l.peekRune(); isHorizontalWhitespace(pr); pr = l.peekRune() {
		l.advance()
	}
}

func isHorizontalWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

func (l *Lexer) readWhile(pred func(rune) bool) string {
	pos := l.pos
	for pr := l.peekRune(); pred(pr); pr = l.peekRune() {
		l.advance()
	}
	return string(l.input[pos : l.pos+1])
}

func (l *Lexer) readComment() string {
	return l.readWhile(func(r rune) bool { return r != 0 && r != '\n' })
}

func (l *Lexer) readNum() string {
	return l.readWhile(func(r rune) bool { return isDigit(r) || r == '.' })
}

func (l *Lexer) readIdent() string {
	return l.readWhile(func(r rune) bool { return isLetter(r) || unicode.IsDigit(r) })
}

func (l *Lexer) readString() (string, error) {
	pos := l.pos
	backslashCnt := 0
	for {
		l.advance()
		if l.cur == '\\' {
			backslashCnt++
		} else {
			backslashCnt = 0
		}
		pr := l.peekRune()
		if pr == '"' && backslashCnt%2 == 0 {
			l.advance() // end of string
			break
		}
		if pr == 0 || pr == '\n' {
			break // error case
		}
	}
	s := string(l.input[pos : l.pos+1])
	r, err := strconv.Unquote(s)
	if err != nil {
		return "", err
	}

	return r, nil
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
