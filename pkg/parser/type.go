package parser

import "foxygo.at/evy/pkg/lexer"

type TypeName int

const (
	ILLEGAL TypeName = iota
	NUM
	STRING
	BOOL
	ANY
	ARRAY
	MAP
	NONE // for functions without return value, declaration statements, etc.
)

var (
	ILLEGAL_TYPE = &Type{Name: ILLEGAL}
	NUM_TYPE     = &Type{Name: NUM}
	BOOL_TYPE    = &Type{Name: BOOL}
	STRING_TYPE  = &Type{Name: STRING}
	ANY_TYPE     = &Type{Name: ANY}
	NONE_TYPE    = &Type{Name: NONE}
)

func isBasicType(t lexer.TokenType) bool {
	return t == lexer.NUM || t == lexer.STRING || t == lexer.BOOL || t == lexer.ANY
}

func basicTypeName(t lexer.TokenType) TypeName {
	switch t {
	case lexer.NUM:
		return NUM
	case lexer.STRING:
		return STRING
	case lexer.BOOL:
		return BOOL
	case lexer.ANY:
		return ANY
	}
	return ILLEGAL
}

func compositeTypeName(t lexer.TokenType) TypeName {
	switch t {
	case lexer.LBRACKET:
		return ARRAY
	case lexer.LCURLY:
		return MAP
	}
	return ILLEGAL
}

var typeNameStrings = map[TypeName]string{
	ILLEGAL: "ILLEGAL",
	NUM:     "NUM",
	STRING:  "STRING",
	BOOL:    "BOOL",
	ANY:     "ANY",
	ARRAY:   "ARRAY",
	MAP:     "MAP",
}

func (t TypeName) String() string {
	if s, ok := typeNameStrings[t]; ok {
		return s
	}
	return "UNKNOWN"
}

func (t TypeName) GoString() string {
	return t.String()
}

type Type struct {
	Name TypeName // string, num, bool, composite types array, map
	Sub  *Type    // e.g.: `[]int` : Type{Name: "array", Sub: &Type{Name: "int"} }
}

func (t *Type) String() string {
	if t.Sub == nil {
		return t.Name.String()
	}
	return t.Name.String() + " " + t.Sub.String()
}

func (t *Type) Accepts(t2 *Type) bool {
	if t.acceptsStrict(t2) {
		return true
	}
	n, n2 := t.Name, t2.Name
	if n == ANY && n2 != ILLEGAL && n2 != NONE {
		return true
	}
	return false
}

// any[] (ARRAY ANY) DOES NOT accept num[] (ARRAY NUM)
func (t *Type) acceptsStrict(t2 *Type) bool {
	n, n2 := t.Name, t2.Name
	if n == ILLEGAL || n == NONE || n2 == ILLEGAL || n2 == NONE {
		return false
	}
	if n != n2 {
		return false
	}
	if t.Sub == nil || t2.Sub == nil {
		return t.Sub == nil && t2.Sub == nil
	}
	return t.Sub.acceptsStrict(t2.Sub)
}
