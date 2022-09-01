package parser

import "foxygo.at/evy/pkg/lexer"

type Type int

const (
	ILLEGAL_TYPE Type = iota
	NUM
	STRING
	BOOL
	ANY
	ARRAY
	MAP
)

func typeFromToken(t lexer.TokenType) Type {
	switch t {
	case lexer.NUM:
		return NUM
	case lexer.STRING:
		return STRING
	case lexer.BOOL:
		return BOOL
	case lexer.ANY:
		return ANY
	case lexer.LBRACKET:
		return ARRAY
	case lexer.LCURLY:
		return MAP
	}
	return ILLEGAL_TYPE
}

func typeFromLiteral(t lexer.TokenType) Type {
	switch t {
	case lexer.NUM_LIT:
		return NUM
	case lexer.STRING_LIT:
		return STRING
	case lexer.TRUE, lexer.FALSE:
		return BOOL
	}
	return ILLEGAL_TYPE
}

var typeStrings = map[Type]string{
	ILLEGAL_TYPE: "ILLEGAL_TYPE",
	NUM:          "NUM",
	STRING:       "STRING",
	BOOL:         "BOOL",
	ANY:          "ANY",
	ARRAY:        "ARRAY",
	MAP:          "MAP",
}

func (t Type) String() string {
	if s, ok := typeStrings[t]; ok {
		return s
	}
	return "UNKNOWN"
}

func (t Type) GoString() string {
	return t.String()
}
