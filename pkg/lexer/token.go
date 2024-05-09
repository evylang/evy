package lexer

import (
	"fmt"
	"strconv"
)

// Token contains
//   - type of the token, such as [IDENT], [PLUS] or [NUM_LIT]
//   - start location of the token in the input string
//   - literal value of the token, used only for number literals, string
//     literals and comments.
type Token struct {
	Literal string

	Offset int
	Line   int
	Col    int
	Type   TokenType
}

// TokenType represents the type of token, such as identifier [IDENT],
// operator [PLUS] or literal [NUM_LIT].
type TokenType int

// Token types are represented as constants and are the core field of
// the [Token] struct type.
const (
	ILLEGAL TokenType = iota
	EOF
	COMMENT // `// a comment`

	// Identifiers and Literals.
	IDENT      // some_identifier
	NUM_LIT    // 123 or 456.78
	STRING_LIT // "a string 🧵"

	// Operators.
	DECLARE  // :=
	ASSIGN   // =
	PLUS     // +
	MINUS    // -
	BANG     // !
	ASTERISK // *
	SLASH    // /
	PERCENT  // %

	EQ     // ==
	NOT_EQ // !=
	LT     // <
	GT     // >
	LTEQ   // <=
	GTEQ   // >=

	// Delimiters.
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	LCURLY   // {
	RCURLY   // }

	COLON // :
	WS    // ' '
	NL    // '\n'
	DOT   // .
	DOT3  // ...

	// Keywords.
	NUM    // num
	STRING // string
	BOOL   // bool
	ANY    // any

	TRUE  // true
	FALSE // false
	AND   // and
	OR    // or

	IF     // if
	ELSE   // else
	FUNC   // func
	RETURN // return
	ON     // on
	FOR    // for
	RANGE  // range
	WHILE  // while
	BREAK  // break
	END    // end
)

func lookupKeyword(s string) TokenType {
	if tok, ok := keywords[s]; ok {
		return tok
	}
	return IDENT
}

var keywords = map[string]TokenType{
	"true":  TRUE,
	"false": FALSE,
	"and":   AND,
	"or":    OR,

	"num":    NUM,
	"string": STRING,
	"bool":   BOOL,
	"any":    ANY,

	"if":     IF,
	"else":   ELSE,
	"func":   FUNC,
	"on":     ON,
	"return": RETURN,
	"for":    FOR,
	"range":  RANGE,
	"while":  WHILE,
	"break":  BREAK,
	"end":    END,
}

// AsIdent returns t as an IDENT token if t is a keyword and valid as an
// identifier, otherwise it returns t. This is to allow specific tokens that
// are also valid identifiers to be used in certain contexts.
func (t *Token) AsIdent() *Token {
	tokstr := tokenStrings[t.TokenType()].format
	if _, ok := keywords[tokstr]; !ok {
		return t
	}
	ident := *t
	ident.setLiteral(tokstr).setType(IDENT)
	return &ident
}

type tokenString struct {
	string string
	format string
}

var tokenStrings = map[TokenType]tokenString{
	ILLEGAL:    {string: "ILLEGAL", format: "💣"},
	EOF:        {string: "EOF", format: ""},
	COMMENT:    {string: "COMMENT", format: ""},
	IDENT:      {string: "IDENT", format: ""},
	NUM_LIT:    {string: "NUM_LIT", format: ""},
	STRING_LIT: {string: "STRING_LIT", format: ""},
	DECLARE:    {string: "DECLARE", format: ":="},
	ASSIGN:     {string: "ASSIGN", format: "="},
	PLUS:       {string: "PLUS", format: "+"},
	MINUS:      {string: "MINUS", format: "-"},
	BANG:       {string: "BANG", format: "!"},
	ASTERISK:   {string: "ASTERISK", format: "*"},
	PERCENT:    {string: "PERCENT", format: "%"},
	SLASH:      {string: "SLASH", format: "/"},
	LT:         {string: "LT", format: "<"},
	GT:         {string: "GT", format: ">"},
	LTEQ:       {string: "LTEQ", format: "<="},
	GTEQ:       {string: "GTEQ", format: ">="},
	EQ:         {string: "EQ", format: "=="},
	NOT_EQ:     {string: "NOT_EQ", format: "!="},
	COLON:      {string: "COLON", format: ":"},
	NL:         {string: "NL", format: "\n"},
	WS:         {string: "WS", format: " "},
	LPAREN:     {string: "LPAREN", format: "("},
	RPAREN:     {string: "RPAREN", format: ")"},
	LBRACKET:   {string: "LBRACKET", format: "["},
	RBRACKET:   {string: "RBRACKET", format: "]"},
	LCURLY:     {string: "LCURLY", format: "{"},
	RCURLY:     {string: "RCURLY", format: "}"},
	DOT:        {string: "DOT", format: "."},
	DOT3:       {string: "DOT3", format: "..."},
	TRUE:       {string: "TRUE", format: "true"},
	FALSE:      {string: "FALSE", format: "false"},
	AND:        {string: "AND", format: "and"},
	OR:         {string: "OR", format: "or"},
	NUM:        {string: "NUM", format: "num"},
	STRING:     {string: "STRING", format: "string"},
	BOOL:       {string: "BOOL", format: "bool"},
	ANY:        {string: "ANY", format: "any"},
	IF:         {string: "IF", format: "if"},
	ELSE:       {string: "ELSE", format: "else"},
	FUNC:       {string: "FUNC", format: "func"},
	RETURN:     {string: "RETURN", format: "return"},
	ON:         {string: "ON", format: "on"},
	FOR:        {string: "FOR", format: "for"},
	RANGE:      {string: "RANGE", format: "range"},
	WHILE:      {string: "WHILE", format: "while"},
	BREAK:      {string: "BREAK", format: "break"},
	END:        {string: "END", format: "end"},
}

// String implements the [fmt.Stringer] interface for [TokenType].
func (t TokenType) String() string {
	if ts, ok := tokenStrings[t]; ok {
		return ts.string
	}
	return "UNKNOWN"
}

// GoString implements the [fmt.GoStringer] interface for
// [TokenType]. Its return value is more useful than the iota value
// when debugging failed tests.
func (t TokenType) GoString() string {
	return t.String()
}

// Format returns a string representation of the token type that is
// useful in error messages. The string representation is more
// descriptive than the string returned by the String() method.
func (t TokenType) Format() string {
	if t == EOF {
		return "end of input"
	}
	if t == NL {
		return "end of line"
	}
	if t == IDENT {
		return "identifier"
	}
	return fmt.Sprintf("%q", tokenStrings[t].format)
}

func (t *Token) setType(tokenType TokenType) *Token {
	t.Type = tokenType
	return t
}

// TokenType returns the type of the token as represented by the
// [TokenType] type.
func (t *Token) TokenType() TokenType {
	return t.Type
}

func (t *Token) setLiteral(literal string) *Token {
	t.Literal = literal
	return t
}

// String implements the [fmt.Stringer] interface for [Token].
func (t *Token) String() string {
	switch t.Type {
	case COMMENT, IDENT, NUM_LIT, STRING_LIT:
		return fmt.Sprintf("%s %q", t.Type, t.Literal)
	case ILLEGAL:
		return fmt.Sprintf("ILLEGAL 💥 %q at %s", t.Literal, t.Location())
	}
	return t.Type.String()
}

// Format returns a string representation of the token that is useful in
// error messages. If the token has a relevant literal value, the
// literal is returned. Otherwise, the format of the token type is
// returned.
func (t *Token) Format() string {
	switch t.Type {
	case COMMENT, IDENT, NUM_LIT:
		return t.Literal
	case STRING_LIT:
		return `"` + t.Literal + `"`
	}
	return t.Type.Format()
}

// Location returns a string representation of a token's start location
// in the form of: "line <line number> column <column number>".
func (t *Token) Location() string {
	return "line " + strconv.Itoa(t.Line) + " column " + strconv.Itoa(t.Col)
}
