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
	ILLEGAL_TYPE  = &Type{Name: ILLEGAL}
	NUM_TYPE      = &Type{Name: NUM}
	BOOL_TYPE     = &Type{Name: BOOL}
	STRING_TYPE   = &Type{Name: STRING}
	ANY_TYPE      = &Type{Name: ANY}
	NONE_TYPE     = &Type{Name: NONE}
	GENERIC_ARRAY = &Type{Name: ARRAY, Sub: NONE_TYPE}
	GENERIC_MAP   = &Type{Name: MAP, Sub: NONE_TYPE}
)

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
	name   string
	format string
}

var typeNameStrings = map[TypeName]typeNameString{
	ILLEGAL: {name: "ILLEGAL", format: "ILLEGAL"},
	NUM:     {name: "num", format: "num"},
	STRING:  {name: "string", format: "string"},
	BOOL:    {name: "bool", format: "bool"},
	ANY:     {name: "any", format: "any"},
	ARRAY:   {name: "array", format: "[]"},
	MAP:     {name: "map", format: "{}"},
	NONE:    {name: "none", format: "none"},
}

func (t TypeName) String() string {
	return typeNameStrings[t].format
}

func (t TypeName) Name() string {
	return typeNameStrings[t].name
}

func (t TypeName) GoString() string {
	return t.String()
}

type Type struct {
	Name TypeName // string, num, bool, composite types array, map
	Sub  *Type    // e.g.: `[]int` : Type{Name: "array", Sub: &Type{Name: "int"} }
}

func (t *Type) String() string {
	if t.Sub == nil || t == GENERIC_ARRAY || t == GENERIC_MAP {
		return t.Name.String()
	}
	return t.Name.String() + t.Sub.String()
}

func (t *Type) Accepts(t2 *Type) bool {
	if t.acceptsStrict(t2) {
		return true
	}
	n, n2 := t.Name, t2.Name
	if n == ANY && n2 != ILLEGAL && n2 != NONE {
		return true
	}
	// empty Array none array accepted by all arrays.
	// empty Map of none_type accepted by all maps
	return false
}

func (t *Type) Matches(t2 *Type) bool {
	if t == t2 {
		return true
	}
	if t.Name != t2.Name {
		return false
	}
	if t.Sub == t2.Sub {
		return true
	}
	if t.Sub == nil || t2.Sub == nil {
		return false
	}

	if t == GENERIC_ARRAY || t == GENERIC_MAP || t2 == GENERIC_ARRAY || t2 == GENERIC_MAP {
		return true
	}
	return t.Sub.Matches(t2.Sub)
}

func (t *Type) Infer() *Type {
	if t.Name != ARRAY && t.Name != MAP {
		return t
	}
	if t.Sub == NONE_TYPE {
		t2 := *t
		t2.Sub = ANY_TYPE
		return &t2
	}
	t.Sub = t.Sub.Infer()
	return t
}

// []any (ARRAY ANY) DOES NOT accept []num (ARRAY NUM).
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
	// all array types except empty array literal
	if n == ARRAY && t2 == GENERIC_ARRAY {
		return true
	}
	// all map types except empty map literal
	if n == MAP && t2 == GENERIC_MAP {
		return true
	}
	return t.Sub.acceptsStrict(t2.Sub)
}

func (t *Type) sameComposite(t2 *Type) bool {
	if t.Name == ARRAY && t2.Name == ARRAY {
		return true
	}
	if t.Name == MAP && t2.Name == MAP {
		return true
	}
	return false
}
