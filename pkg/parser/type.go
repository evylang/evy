package parser

import (
	"foxygo.at/evy/pkg/lexer"
)

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

func compositeTypeName(t lexer.TokenType) TypeName {
	switch t {
	case lexer.LBRACKET:
		return ARRAY
	case lexer.LCURLY:
		return MAP
	}
	return ILLEGAL
}

type typeNameString struct {
	string string
	format string
}

var typeNameStrings = map[TypeName]typeNameString{
	ILLEGAL: {string: "ILLEGAL", format: "ILLEGAL"},
	NUM:     {string: "NUM", format: "num"},
	STRING:  {string: "STRING", format: "string"},
	BOOL:    {string: "BOOL", format: "bool"},
	ANY:     {string: "ANY", format: "any"},
	ARRAY:   {string: "ARRAY", format: "[]"},
	MAP:     {string: "MAP", format: "{}"},
	NONE:    {string: "NONE", format: "none"},
}

func (t TypeName) String() string {
	if s, ok := typeNameStrings[t]; ok {
		return s.string
	}
	return "UNKNOWN"
}

func (t TypeName) Format() string {
	if s, ok := typeNameStrings[t]; ok {
		return s.format
	}
	return "<unknown>"
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

func (t *Type) Format() string {
	if t.Sub == nil {
		return t.Name.Format()
	}
	return t.Sub.Format() + t.Name.Format()
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
	if n == ILLEGAL || n2 == ILLEGAL {
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
