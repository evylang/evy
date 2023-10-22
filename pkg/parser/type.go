package parser

import (
	"evylang.dev/evy/pkg/lexer"
)

// TypeName represents the enumerable basic types (such as num, string,
// and bool), categories of composite types (such as array and map),
// the dynamic type (any), and the none type, which is used where no
// type is expected. TypeName is used in the [Type] struct which fully
// specifies all types, including composite types.
type TypeName int

// The enumerated basic types, categories of composite types, any and
// none are defined as constants.
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

// Basic types, any, none, and untyped arrays and untyped maps are
// [interned] into variables for reuse, such as [NUM_TYPE] or
// [UNTYPED_MAP].
//
// [interned]: https://en.wikipedia.org/wiki/Interning_(computer_science)
var (
	ILLEGAL_TYPE  = &Type{Name: ILLEGAL}
	NUM_TYPE      = &Type{Name: NUM}
	BOOL_TYPE     = &Type{Name: BOOL}
	STRING_TYPE   = &Type{Name: STRING}
	ANY_TYPE      = &Type{Name: ANY}
	NONE_TYPE     = &Type{Name: NONE}
	UNTYPED_ARRAY = &Type{Name: ARRAY, Sub: NONE_TYPE}
	UNTYPED_MAP   = &Type{Name: MAP, Sub: NONE_TYPE}
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

func (t TypeName) name() string {
	return typeNameStrings[t].name
}

// Type holds a full representation of any Evy variable or value. It can
// represent basic types, such as numbers and strings, as well as
// composite types, such as arrays and maps. It is also used to
// represent the dynamic [ANY_TYPE]. For AST nodes that have no type
// [NONE_TYPE] is used.
type Type struct {
	Name TypeName // string, num, bool, composite types array, map
	Sub  *Type    // e.g.: `[]int` : Type{Name: "array", Sub: &Type{Name: "int"} }
}

func (t *Type) String() string {
	if t.Sub == nil || t == UNTYPED_ARRAY || t == UNTYPED_MAP {
		return t.Name.String()
	}
	return t.Name.String() + t.Sub.String()
}

func (t *Type) accepts(t2 *Type) bool {
	if t.acceptsStrict(t2) {
		return true
	}
	n, n2 := t.Name, t2.Name
	if n == ANY && n2 != ILLEGAL && n2 != NONE {
		return true
	}
	// empty Array of none_type accepted by all arrays.
	// empty Map of none_type accepted by all maps
	return false
}

// Matches returns true if the two types are equal, or if one is a
// untyped array and the other is a specific array, or if one is a
// untyped map and the other is a specific map. This is useful for type
// validation in binary expressions, such as array concatenation:
//
//	[] + [1]
//	[1] + []
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
	if t == UNTYPED_ARRAY || t == UNTYPED_MAP || t2 == UNTYPED_ARRAY || t2 == UNTYPED_MAP {
		return true
	}
	return t.Sub.Matches(t2.Sub)
}

func (t *Type) infer() *Type {
	if t.Name != ARRAY && t.Name != MAP {
		return t
	}
	if t.Sub == NONE_TYPE {
		t2 := *t
		t2.Sub = ANY_TYPE
		return &t2
	}
	t.Sub = t.Sub.infer()
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
	if n == ARRAY && t2 == UNTYPED_ARRAY {
		return true
	}
	// all map types except empty map literal
	if n == MAP && t2 == UNTYPED_MAP {
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

func combineTypes(types []*Type) *Type {
	combinedT := types[0]
	for _, t := range types[1:] {
		if combinedT.accepts(t) {
			continue
		}
		if t.accepts(combinedT) {
			combinedT = t
			continue
		}
		if t.sameComposite(combinedT) {
			combinedT = &Type{Name: t.Name, Sub: ANY_TYPE}
			continue
		}
		return ANY_TYPE
	}
	return combinedT
}
