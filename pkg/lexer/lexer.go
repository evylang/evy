// Package lexer tokenizes the input.
//
// The first step in the Evy compilation process is tokenization. This
// involves breaking the Evy input code into individual tokens, such as
// keywords, operators, and identifiers.  The lexer package is
// responsible for this task. It provides a [Lexer] type, which can be
// initialized using the [New] function. The [Lexer.Next] method
// returns the next Token in the input string. A [Token] is a data
// structure that represents a single token in the input code. The
// [EOF] token is a special token that indicates the end of the input
// code.
//
// The [parser] then takes these tokens and creates an Abstract Syntax
// Tree(AST), which is a representation of the Evy code's structure.
// Finally, the [evaluator] walks the AST and executes the program.
//
// [parser]: https://pkg.go.dev/foxygo.at/evy/pkg/parser
// [evaluator]: https://pkg.go.dev/foxygo.at/evy/pkg/evaluator
package lexer

import (
	"strconv"
	"unicode"
)

// Lexer is a lexical analyzer for Evy source code.
type Lexer struct {
	input []rune
	cur   rune // current rune under examination
	pos   int  // current position in input (points to current rune)
	line  int
	col   int
}

// New creates a new Lexer for the given input string.
func New(input string) *Lexer {
	return &Lexer{input: []rune(input), pos: -1, line: 1}
}

// Next returns the next [Token] in the input string. When the end of
// the input string is reached Next returns a Token with type [EOF].
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
		return tok.setType(WS)
	case '=':
		if l.peekRune() == '=' {
			l.advance()
			return tok.setType(EQ)
		}
		return tok.setType(ASSIGN)
	case '+':
		return tok.setType(PLUS)
	case '-':
		return tok.setType(MINUS)
	case '!':
		if l.peekRune() == '=' {
			l.advance()
			return tok.setType(NOT_EQ)
		}
		return tok.setType(BANG)
	case '/':
		if l.peekRune() == '/' {
			return tok.setType(COMMENT).setLiteral(l.readComment())
		}
		return tok.setType(SLASH)
	case '*':
		return tok.setType(ASTERISK)
	case '%':
		return tok.setType(PERCENT)
	case '<':
		if l.peekRune() == '=' {
			l.advance()
			return tok.setType(LTEQ)
		}
		return tok.setType(LT)
	case '>':
		if l.peekRune() == '=' {
			l.advance()
			return tok.setType(GTEQ)
		}
		return tok.setType(GT)
	case ':':
		if l.peekRune() == '=' {
			l.advance()
			return tok.setType(DECLARE)
		}
		return tok.setType(COLON)
	case '{':
		return tok.setType(LCURLY)
	case '}':
		return tok.setType(RCURLY)
	case '(':
		return tok.setType(LPAREN)
	case ')':
		return tok.setType(RPAREN)
	case '[':
		return tok.setType(LBRACKET)
	case ']':
		return tok.setType(RBRACKET)
	case '\n':
		return tok.setType(NL)
	case '.':
		if l.peekRune() == '.' && l.peekRune2() == '.' {
			l.advance()
			l.advance()
			return tok.setType(DOT3)
		}
		return tok.setType(DOT)
	case '"':
		literal, err := l.readString()
		// strconv.Unquote error
		if err != nil {
			return tok.setType(ILLEGAL).setLiteral("invalid string")
		}
		return tok.setType(STRING_LIT).setLiteral(literal)
	case 0:
		return tok.setType(EOF)
	}
	if isLetter(l.cur) {
		literal := l.readIdent()
		tokenType := lookupKeyword(literal)
		if tokenType == IDENT {
			return tok.setType(IDENT).setLiteral(literal)
		}
		return tok.setType(tokenType)
	}
	if isDigit(l.cur) {
		return tok.setType(NUM_LIT).setLiteral(l.readNum())
	}

	return tok.setType(ILLEGAL).setLiteral(string(l.cur))
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
	escaped := false
	for {
		escaped = l.cur == '\\' && !escaped
		pr := l.peekRune()
		if pr == '"' && !escaped {
			l.advance() // end of string
			break
		}
		if pr == 0 || pr == '\n' {
			break // error case
		}
		l.advance()
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
