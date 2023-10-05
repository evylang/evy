package parser

import "evylang.dev/evy/pkg/lexer"

// Operator represents the operators used in binary and unary
// expressions. For example, the OP_ASTERISK operator represents the
// multiplication operator (*).
type Operator int

// Operators are represented as constants and are used in
// [BinaryExpression] and [UnaryExpression].
const (
	OP_ILLEGAL Operator = iota
	OP_PLUS
	OP_MINUS
	OP_SLASH
	OP_ASTERISK
	OP_PERCENT

	OP_OR
	OP_AND

	OP_EQ
	OP_NOT_EQ
	OP_LT
	OP_GT
	OP_LTEQ
	OP_GTEQ

	OP_INDEX
	OP_DOT
	OP_BANG
)

var operatorStrings = map[Operator]string{
	OP_ILLEGAL:  "illegal",
	OP_PLUS:     "+",
	OP_MINUS:    "-",
	OP_SLASH:    "/",
	OP_ASTERISK: "*",
	OP_PERCENT:  "%",
	OP_OR:       "or",
	OP_AND:      "and",
	OP_EQ:       "==",
	OP_NOT_EQ:   "!=",
	OP_LT:       "<",
	OP_GT:       ">",
	OP_LTEQ:     "<=",
	OP_GTEQ:     ">=",
	OP_INDEX:    "[op_index]",
	OP_DOT:      ".",
	OP_BANG:     "!",
}

func op(tok *lexer.Token) Operator {
	switch tok.Type {
	case lexer.PLUS:
		return OP_PLUS
	case lexer.MINUS:
		return OP_MINUS
	case lexer.SLASH:
		return OP_SLASH
	case lexer.ASTERISK:
		return OP_ASTERISK
	case lexer.PERCENT:
		return OP_PERCENT
	case lexer.OR:
		return OP_OR
	case lexer.AND:
		return OP_AND
	case lexer.EQ:
		return OP_EQ
	case lexer.NOT_EQ:
		return OP_NOT_EQ
	case lexer.LT:
		return OP_LT
	case lexer.GT:
		return OP_GT
	case lexer.LTEQ:
		return OP_LTEQ
	case lexer.GTEQ:
		return OP_GTEQ
	case lexer.LBRACKET:
		return OP_INDEX
	case lexer.DOT:
		return OP_DOT
	case lexer.BANG:
		return OP_BANG
	}
	return OP_ILLEGAL
}

// String returns a string representation of the operator, such as "+"
// for OP_PLUS.
func (o Operator) String() string {
	return operatorStrings[o]
}
