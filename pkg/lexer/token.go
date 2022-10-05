package lexer

import (
	"strconv"
)

type Token struct {
	Literal string

	Offset int
	Line   int
	Col    int
	Type   TokenType
}

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	COMMENT // `// a comment`

	// Identifiers and Literals.
	IDENT      // some_identifier
	NUM_LIT    // 123 456.78
	STRING_LIT // "a string 🧵"

	// Operators.
	DECLARE  // :=
	ASSIGN   // =
	PLUS     // +
	MINUS    // -
	BANG     // !
	ASTERISK // *
	SLASH    // /

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

func LookupKeyword(s string) TokenType {
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
	STRING_LIT: {string: "STRING_LIT", format: "string"},
	DECLARE:    {string: "DECLARE", format: ":="},
	ASSIGN:     {string: "ASSIGN", format: "="},
	PLUS:       {string: "PLUS", format: "+"},
	MINUS:      {string: "MINUS", format: "-"},
	BANG:       {string: "BANG", format: "!"},
	ASTERISK:   {string: "ASTERISK", format: "*"},
	SLASH:      {string: "SLASH", format: "/"},
	LT:         {string: "LT", format: "<"},
	GT:         {string: "GT", format: ">"},
	LTEQ:       {string: "LTEQ", format: "<="},
	GTEQ:       {string: "GTEQ", format: ">="},
	EQ:         {string: "EQ", format: "=="},
	NOT_EQ:     {string: "NOT_EQ", format: "!="},
	COLON:      {string: "COLON", format: ":"},
	NL:         {string: "NL", format: "\n"},
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

func (t TokenType) Format() string {
	if ts, ok := tokenStrings[t]; ok {
		return ts.format
	}
	return "<unknown>"
}

func (t TokenType) String() string {
	if ts, ok := tokenStrings[t]; ok {
		return ts.string
	}
	return "UNKNOWN"
}

func (t TokenType) GoString() string {
	return t.String()
}

func (t TokenType) FormatDetails() string {
	if t == EOF {
		return "end of input"
	}
	if t == NL {
		return "end of line"
	}
	return "'" + t.Format() + "'"
}

func (t *Token) SetType(tokenType TokenType) *Token {
	t.Type = tokenType
	return t
}

func (t *Token) TokenType() TokenType {
	return t.Type
}

func (t *Token) SetLiteral(literal string) *Token {
	t.Literal = literal
	return t
}

func (t *Token) String() string {
	switch t.Type {
	case COMMENT, IDENT, NUM_LIT, STRING_LIT:
		return t.Type.String() + " '" + t.Literal + "'"
	case ILLEGAL:
		return "ILLEGAL 💥 '" + t.Literal + "' at " + t.Location()
	}
	return t.Type.String()
}

func (t *Token) Format() string {
	switch t.Type {
	case COMMENT, IDENT, NUM_LIT:
		return t.Literal
	case STRING_LIT:
		return `"` + t.Literal + `"`
	}
	return t.Type.Format()
}

func (t *Token) FormatDetails() string {
	switch t.Type {
	case COMMENT, IDENT, NUM_LIT:
		return t.Literal
	case STRING_LIT:
		return `"` + t.Literal + `"`
	}
	return t.Type.FormatDetails()
}

func (t *Token) Location() string {
	return "line " + strconv.Itoa(t.Line) + " column " + strconv.Itoa(t.Col)
}
