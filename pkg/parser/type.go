package parser

import (
	"foxygo.at/evy/pkg/lexer"
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

// Equals returns true if Type t equals to t2, including subtypes.
func (t *Type) Equals(t2 *Type) bool {
	if t == t2 {
		return true
	}
	if t.Name != t2.Name {
		return false
	}
	if t.Sub == nil || t2.Sub == nil {
		return t.Sub == nil && t2.Sub == nil
	}
	return t.Sub.Equals(t2.Sub)
}

func (t *Type) accepts(t2 *Type) bool {
	return t.matches(t2) || t == ANY_TYPE
}

// Matches returns true if the two types are equal, or if one is a
// untyped array and the other is a specific array, or if one is a
// untyped map and the other is a specific map. This is useful for type
// validation in binary expressions, such as array concatenation:
//
//	[] + [1]
//	[1] + []
//
// Additionally some built-in functions make use of the UNTYPED_MAP and
// UNTYPED_ARRAY type in its parameters to implement generic behavior,
// such as in the evy built-in functions
//
//	// UNTYPED_MAP builtin
//	has {a:1} "a" // true
//
//	// UNTYPED_ARRAY builtin
//	join [1 2 3] "." // "1.2.3"
func (t *Type) matches(t2 *Type) bool {
	if t == t2 {
		return true
	}
	if t.Name != t2.Name {
		return false
	}
	if t.Sub == nil || t2.Sub == nil {
		return t.Sub == nil && t2.Sub == nil
	}
	if t == UNTYPED_ARRAY || t == UNTYPED_MAP || t2 == UNTYPED_ARRAY || t2 == UNTYPED_MAP {
		return true
	}
	return t.Sub.matches(t2.Sub)
}

// Infer returns the default inferred type of an empty composite type.
// For example, [] becomes []any, and {{}} becomes {}{}any.
func (t *Type) Infer() *Type {
	if t.Name != ARRAY && t.Name != MAP {
		return t
	}
	// empty array becomes []any, empty map becomes {}any
	if t == UNTYPED_ARRAY || t == UNTYPED_MAP {
		t2 := *t
		t2.Sub = ANY_TYPE
		return &t2
	}
	t.Sub = t.Sub.Infer()
	return t
}

// IsUntyped returns true if the type is an untyped array, an untyped map
// or a combination. It returns true for the following literals:
//
// [], {}, [[]], [{}]
//
// However, it returns false for a variable of type []any or [[] {}].
func (t *Type) IsUntyped() bool {
	if t.Name != ARRAY && t.Name != MAP {
		return false
	}
	if t.Sub == NONE_TYPE {
		return true
	}
	return t.Sub.IsUntyped()
}
