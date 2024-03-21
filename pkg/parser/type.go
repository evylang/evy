package parser

// TypeName represents the enumerable basic types (such as num, string,
// and bool), categories of composite types (such as array and map),
// the dynamic type (any), and the none type, which is used where no
// type is expected. TypeName is used in the [Type] struct which fully
// specifies all types, including composite types.
type TypeName int

// The enumerated basic types, categories of composite types, any and
// none are defined as constants.
const (
	NUM TypeName = iota
	STRING
	BOOL
	ANY
	ARRAY
	MAP
	NONE // for functions without return value, declaration statements, etc.
)

// Basic types, any, none, and untyped arrays and untyped maps are
// [interned] into variables for reuse, such as [NUM_TYPE] or
// [EMPTY_MAP].
//
// [interned]: https://en.wikipedia.org/wiki/Interning_(computer_science)
var (
	NUM_TYPE      = &Type{Name: NUM}
	BOOL_TYPE     = &Type{Name: BOOL}
	STRING_TYPE   = &Type{Name: STRING}
	ANY_TYPE      = &Type{Name: ANY}
	NONE_TYPE     = &Type{Name: NONE}
	EMPTY_ARRAY   = &Type{Name: ARRAY, Sub: NONE_TYPE}
	EMPTY_MAP     = &Type{Name: MAP, Sub: NONE_TYPE}
	GENERIC_ARRAY = &Type{Name: ARRAY}
	GENERIC_MAP   = &Type{Name: MAP}
)

type typeNameString struct {
	name   string
	format string
}

var typeNameStrings = map[TypeName]typeNameString{
	NUM:    {name: "num", format: "num"},
	STRING: {name: "string", format: "string"},
	BOOL:   {name: "bool", format: "bool"},
	ANY:    {name: "any", format: "any"},
	ARRAY:  {name: "array", format: "[]"},
	MAP:    {name: "map", format: "{}"},
	NONE:   {name: "none", format: "none"},
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

	// Fixed is a flag relevant only to composite types: arrays and maps. It
	// determines if the composite type can be converted (or coerced) to a
	// different composite type.
	//
	// *  Fixed is false: The composite type is flexible. It can be coerced
	//    to other composite types (e.g., []num to []any) or directly to the
	//    dynamic `any` type. This applies to composite literals and nested
	//    composite literals.
	// *  Fixed is true: The composite type is strict. It cannot be directly
	//    coerced to other composite types, only to the dynamic `any` type.
	//    This applies to variables and expressions (like array
	//    concatenations and slices).
	Fixed bool
}

func fixedType(t *Type) *Type {
	if t.Name != ARRAY && t.Name != MAP {
		return t
	}
	if t == GENERIC_ARRAY || t == GENERIC_MAP || t == EMPTY_ARRAY || t == EMPTY_MAP {
		return t
	}
	t2 := *t
	t2.Fixed = true
	return &t2
}

// String returns a string representation of the Type.
func (t *Type) String() string {
	if t == nil {
		return "ILLEGAL"
	}
	if t.Sub == nil || t == EMPTY_ARRAY || t == EMPTY_MAP {
		return t.Name.String()
	}
	return t.Name.String() + t.Sub.String()
}

// Equals returns if t and t2 and all their sub types are equal in Name.
func (t *Type) Equals(t2 *Type) bool {
	left, right := t, t2
	for (left != nil) && (right != nil) {
		switch {
		case left == right:
			return true
		case left.Name != right.Name:
			return false
		}
		left, right = left.Sub, right.Sub
	}
	return left == right
}

// accepts reports whether a variable 'v' of Type 't' can be assigned a
// value 'v2' of Type 't2'. accepts returns true if v = v2 is a valid
// assignment.
//
// Assignability of composite types (arrays, maps) is influenced by whether
// the value is a literal or a variable. The internal Type.Fixed flag tracks
// this distinction.
//
// A composite literal with a specific element type (e.g. []num, {}bool) can
// be assigned to a variable with a compatible 'any' element type (e.g.,
// []any, {}any}).
//
// Composite variables have stricter typing. Assignment is generally only
// allowed between variables of the same composite type, or from a composite
// variable to an 'any' typed variable.
//
//	anyArr := &Type{Name ARRAY: Type: ANY} //  []any
//	numArr := &Type{Name ARRAY: Type: NUM} // []num (literal)
//	numArrVar := &Type{Name ARRAY: Type: NUM, Fixed: true} // []num (variable)
//	fmt.Println(anyArray.accepts(numArray)) // true
//	fmt.Println(anyArray.accepts(numArrayVar)) // false
func (t *Type) accepts(t2 *Type) bool {
	left, right := t, t2
	var rightFixed bool
	for (left != nil) && (right != nil) {
		if right.Fixed {
			rightFixed = true // cannot be coerced into different composite from here
		}
		switch {
		case left == right:
			return true
		case left.Name == ANY && right.Name != NONE && (left == t || !rightFixed):
			// left == t allows, e.g., any = num but not []any = []num
			return true
		case left.Name != right.Name:
			return false
		case left == GENERIC_ARRAY, left == GENERIC_MAP:
			// "generic" builtins parameter such as `has` for maps.
			return true
		case right == EMPTY_ARRAY, right == EMPTY_MAP:
			return true
		}
		left, right = left.Sub, right.Sub
	}
	return left == right
}

// matches returns true if the two types are equal, or if one is a untyped
// array and the other is a specific array, or if one is a untyped map and
// the other is a specific map. This is used only for type validation in
// binary expressions, such as array concatenation:
//
//	[] + [1]
//	[1] + []
func (t *Type) matches(t2 *Type) bool {
	left, right := t, t2
	for (left != nil) && (right != nil) {
		switch {
		case left == right:
			return true
		case left.Name != right.Name:
			return false
		case left == EMPTY_ARRAY, left == EMPTY_MAP, right == EMPTY_ARRAY, right == EMPTY_MAP:
			return true
		}
		left, right = left.Sub, right.Sub
	}
	return left == right
}

func (t *Type) infer() *Type {
	if t.Name != ARRAY && t.Name != MAP {
		return t
	}
	if t == EMPTY_ARRAY {
		return &Type{Name: ARRAY, Sub: ANY_TYPE}
	}
	if t == EMPTY_MAP {
		return &Type{Name: MAP, Sub: ANY_TYPE}
	}
	t2 := *t
	t2.Sub = t.Sub.infer()
	return &t2
}

func combineTypes(types []*Type) *Type {
	combinedT := types[0]
	for _, t := range types[1:] {
		if combinedT.Equals(t) {
			continue
		}
		// types are not equal, ensure that composite types can be combined
		if t.Fixed || combinedT.Fixed {
			return ANY_TYPE
		}
		if (t.Name == ARRAY || t.Name == MAP) && t.Name == combinedT.Name {
			switch {
			case t == EMPTY_ARRAY, t == EMPTY_MAP: // do nothing
			case combinedT == EMPTY_ARRAY, combinedT == EMPTY_MAP:
				combinedT = t
			default:
				// Only literal composite types of the same kind (array or
				// map) can be combined.
				sub := combineTypes([]*Type{t.Sub, combinedT.Sub})
				combinedT = &Type{Name: t.Name, Sub: sub}
			}
			continue
		}
		return ANY_TYPE
	}
	return combinedT
}
