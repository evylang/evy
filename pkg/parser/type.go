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

// accepts returns true if a variable v of Type t and a value v2 of Type t2
// can form the valid assignment statement v = v2.
//
// Assignability rules are stricter for composite variables than for composite
// constants. Specifically, an any composite type, such as []any, can accept
// a non-any composite type, such as []num, if the non-any composite type is
// the type of a constant, such as [1 2].
//
//	anyArr := &Type{Name ARRAY: Type: ANY} // []any
//	numArr := &Type{Name ARRAY: Type: NUM} // []num
//	fmt.Println(anyArray.accepts(numArray, true /* constant */)) // true
//	fmt.Println(anyArray.accepts(numArray, false /* constant */)) // false
func (t *Type) accepts(t2 *Type, constant bool) bool {
	left, right := t, t2
	for (left != nil) && (right != nil) {
		switch {
		case left == right:
			return true
		case left.Name == ANY && right.Name != NONE && (left == t || constant):
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
		if (t.Name == ARRAY || t.Name == MAP) && t.Name == combinedT.Name {
			switch {
			case t == EMPTY_ARRAY, t == EMPTY_MAP:
			case combinedT == EMPTY_ARRAY, combinedT == EMPTY_MAP:
				combinedT = t
			case t.Name == ARRAY && combinedT.Name == ARRAY, t.Name == MAP && combinedT.Name == MAP:
				sub := combineTypes([]*Type{t.Sub, combinedT.Sub})
				combinedT = &Type{Name: t.Name, Sub: sub}
			default:
				combinedT = &Type{Name: t.Name, Sub: ANY_TYPE}
			}
			continue
		}
		return ANY_TYPE
	}
	return combinedT
}
