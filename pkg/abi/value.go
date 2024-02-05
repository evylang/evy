package abi

import (
	"fmt"
	"strconv"
	"strings"

	"evylang.dev/evy/pkg/parser"
)

// Value is a representation of literals or variable values during evaluation and compilation stages.
// The [Evaluator] returns a Value for every ast node, very commonly nil. It also tracks the global variables
// the stack of local scopes which are represented as maps of variable name to value, maps[string]Value.
type Value interface {
	// Type is the parser type of the value ie string, num, bool, any or composite types array, map.
	Type() *parser.Type
	Equals(Value) bool
	// String returns a string representation of the value.
	String() string
	// Set assigns the internal representation of this Value to the passed parameter.
	Set(Value)
}

// NumVal represents a numeric value.
type NumVal struct {
	V float64
}

// BoolVal represents a boolean value.
type BoolVal struct {
	V bool
}

// StringVal represents a string value.
type StringVal struct {
	V         string
	runeSlice []rune
}

// AnyVal wraps other values. The contained value V may not by of type AnyVal.
type AnyVal struct {
	V Value
}

// ArrayVal represents an array of values.
type ArrayVal struct {
	Elements *[]Value
	T        *parser.Type
}

// MapVal represents a map of values.
type MapVal struct {
	Pairs map[string]Value
	Order *[]string
	T     *parser.Type
}

// ReturnVal represents a return statement.
type ReturnVal struct {
	V Value
}

// BreakVal represents a break statement.
type BreakVal struct{}

// NoneVal represents the bare return value.
type NoneVal struct{}

// Type is the parser type of this Value.
func (n *NumVal) Type() *parser.Type { return parser.NUM_TYPE }

// String returns a string representation of the Value.
func (n *NumVal) String() string { return strconv.FormatFloat(n.V, 'f', -1, 64) }

// Equals returns true if the provided Value equals this Value.
func (n *NumVal) Equals(v Value) bool {
	n2, ok := v.(*NumVal)
	if !ok {
		panic("internal error: Num.Equals called with non-Num value")
	}
	return n.V == n2.V
}

// Set assigns the internal representation of this Value to the passed parameter.
func (n *NumVal) Set(v Value) {
	n2, ok := v.(*NumVal)
	if !ok {
		panic("internal error: Num.Set called with with non-Num value")
	}
	*n = *n2
}

// Type is the parser type of this Value.
func (s *StringVal) Type() *parser.Type { return parser.STRING_TYPE }

// String returns a string representation of the Value.
func (s *StringVal) String() string { return s.V }

// Equals returns true if the provided Value equals this Value.
func (s *StringVal) Equals(v Value) bool {
	s2, ok := v.(*StringVal)
	if !ok {
		panic("internal error: String.Equals called with non-String value")
	}
	return s.V == s2.V
}

// Set assigns the internal representation of this Value to the passed parameter.
func (s *StringVal) Set(v Value) {
	s2, ok := v.(*StringVal)
	if !ok {
		panic("internal error: String.Set called with with non-String value")
	}
	*s = *s2
}

func (s *StringVal) Runes() []rune {
	if s.runeSlice == nil {
		s.runeSlice = []rune(s.V)
	}
	return s.runeSlice
}

// Index returns the value at index, or an error if the value.
func (s *StringVal) Index(idx Value) (Value, error) {
	runes := s.Runes()
	i, err := normalizeIndex(idx, len(runes))
	if err != nil {
		return nil, err
	}
	return &StringVal{V: string(runes[i])}, nil
}

// Slice returns a subset of this string.
func (s *StringVal) Slice(start, end Value) (Value, error) {
	runes := s.Runes()
	length := len(runes)
	startIdx, endIdx, err := normalizeSliceIndices(start, end, length)
	if err != nil {
		return nil, err
	}
	return &StringVal{V: string(runes[startIdx:endIdx])}, nil
}

// Type is the parser type of this Value.
func (*BoolVal) Type() *parser.Type { return parser.BOOL_TYPE }

// String returns a string representation of the Value.
func (b *BoolVal) String() string {
	return strconv.FormatBool(b.V)
}

// Equals returns true if the provided Value equals this Value.
func (b *BoolVal) Equals(v Value) bool {
	b2, ok := v.(*BoolVal)
	if !ok {
		panic("internal error: Bool.Equals called with non-Bool value")
	}
	return b.V == b2.V
}

// Set assigns the internal representation of this Value to the passed parameter.
func (b *BoolVal) Set(v Value) {
	b2, ok := v.(*BoolVal)
	if !ok {
		panic("internal error: Bool.Set called with with non-Bool value")
	}
	*b = *b2
}

// Type is the parser type of this Value.
func (*AnyVal) Type() *parser.Type { return parser.ANY_TYPE }

// String returns a string representation of the Value.
func (a *AnyVal) String() string {
	return a.V.String()
}

// Equals returns true if the provided Value equals this Value.
func (a *AnyVal) Equals(v Value) bool {
	a2, ok := v.(*AnyVal)
	if !ok {
		panic("internal error: Any.Equals called with non-Any value")
	}
	return a.V.Equals(a2.V)
}

// Set assigns the internal representation of this Value to the passed parameter.
func (a *AnyVal) Set(v Value) {
	if a2, ok := v.(*AnyVal); ok {
		a.V = CopyOrRef(a2.V)
	} else {
		a.V = CopyOrRef(v)
	}
}

// Type is the parser type of this Value.
func (r *ReturnVal) Type() *parser.Type { return r.V.Type() }

// String returns a string representation of the Value.
func (r *ReturnVal) String() string { return r.V.String() }

// Equals returns true if the provided Value equals this Value.
func (r *ReturnVal) Equals(v Value) bool { return r.V.Equals(v) }

// Set assigns the internal representation of this Value to the passed parameter.
func (r *ReturnVal) Set(v Value) { r.V.Set(v) }

// Type is the parser type of this Value.
func (r *BreakVal) Type() *parser.Type { return parser.NONE_TYPE }

// String returns a string representation of the Value.
func (r *BreakVal) String() string { return "" }

// Equals returns true if the provided Value equals this Value.
func (r *BreakVal) Equals(_ Value) bool { return false }

// Set assigns the internal representation of this Value to the passed parameter.
func (r *BreakVal) Set(_ Value) {}

// Type is the parser type of this Value.
func (a *ArrayVal) Type() *parser.Type { return a.T }

// String returns a string representation of the Value.
func (a *ArrayVal) String() string {
	elements := make([]string, len(*a.Elements))
	for i, e := range *a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, " ") + "]"
}

// Equals returns true if the provided Value equals this Value.
func (a *ArrayVal) Equals(v Value) bool {
	a2, ok := v.(*ArrayVal)
	if !ok {
		panic("internal error: Array.Equals called with non-Array value")
	}
	if len(*a.Elements) != len(*a2.Elements) {
		return false
	}
	elements2 := *a2.Elements
	for i, e := range *a.Elements {
		e2 := elements2[i]
		if !e.Equals(e2) {
			return false
		}
	}
	return true
}

// Set assigns the internal representation of this Value to the passed parameter.
func (a *ArrayVal) Set(v Value) {
	a2, ok := v.(*ArrayVal)
	if !ok {
		panic("internal error: Array.Set called with with non-Array value")
	}
	a.Elements = a2.Elements
}

// Index returns the value at the specified index.
func (a *ArrayVal) Index(idx Value) (Value, error) {
	i, err := normalizeIndex(idx, len(*a.Elements))
	if err != nil {
		return nil, err
	}
	elements := *a.Elements
	return elements[i], nil
}

// Copy returns a copy of this array.
func (a *ArrayVal) Copy() *ArrayVal {
	elements := make([]Value, len(*a.Elements))
	for i, v := range *a.Elements {
		elements[i] = CopyOrRef(v)
	}
	return &ArrayVal{Elements: &elements, T: a.T}
}

// Slice returns a subset of this array.
func (a *ArrayVal) Slice(start, end Value) (Value, error) {
	length := len(*a.Elements)
	startIdx, endIdx, err := normalizeSliceIndices(start, end, length)
	if err != nil {
		return nil, err
	}

	elements := make([]Value, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		v := (*a.Elements)[i]
		elements[i-startIdx] = CopyOrRef(v)
	}
	return &ArrayVal{Elements: &elements, T: a.T}, nil
}

// CopyOrRef is a copy of the input value for basic types and a
// reference to the value for composite types (arrays and maps).
func CopyOrRef(val Value) Value {
	switch v := val.(type) {
	case *NumVal:
		return &NumVal{V: v.V}
	case *StringVal:
		return &StringVal{V: v.V}
	case *BoolVal:
		return &BoolVal{V: v.V}
	case *AnyVal:
		return &AnyVal{V: CopyOrRef(v.V)}
	case *ArrayVal:
		return v
	case *MapVal:
		return v
	}
	panic("internal error: CopyOrRef called with with invalid value")
}

// Type is the parser type of this Value.
func (m *MapVal) Type() *parser.Type { return m.T }

// String returns a string representation of the Value.
func (m *MapVal) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for _, key := range *m.Order {
		pairs = append(pairs, key+":"+m.Pairs[key].String())
	}
	return "{" + strings.Join(pairs, " ") + "}"
}

// Equals returns true if the provided Value equals this Value.
func (m *MapVal) Equals(v Value) bool {
	m2, ok := v.(*MapVal)
	if !ok {
		panic("internal error: Map.Equals called with non-Map value")
	}
	if len(m.Pairs) != len(m2.Pairs) {
		return false
	}
	for key, val := range m.Pairs {
		val2 := m2.Pairs[key]
		if val2 == nil || !val.Equals(val2) {
			return false
		}
	}
	return true
}

// Set assigns the internal representation of this Value to the passed parameter.
func (m *MapVal) Set(v Value) {
	m2, ok := v.(*MapVal)
	if !ok {
		panic("internal error: Map.Set called with with non-Map value")
	}
	m.Pairs = m2.Pairs
	m.Order = m2.Order
}

// Get returns the value for the provided key, or
// errors if the key was not found.
func (m *MapVal) Get(key string) (Value, error) {
	val, ok := m.Pairs[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrMapKey, key)
	}
	return val, nil
}

// InsertKey adds the key to the map with a zero value
// of the provided type.
func (m *MapVal) InsertKey(key string, t *parser.Type) {
	if _, ok := m.Pairs[key]; ok {
		return
	}
	*m.Order = append(*m.Order, key)
	m.Pairs[key] = Zero(t)
}

// Delete removes the value associated with the provided
// key from the map.
func (m *MapVal) Delete(key string) {
	if _, ok := m.Pairs[key]; !ok {
		return
	}
	delete(m.Pairs, key)
	for i, k := range *m.Order {
		if k == key {
			*m.Order = append((*m.Order)[:i], (*m.Order)[i+1:]...)
			break
		}
	}
}

func normalizeSliceIndices(start, end Value, length int) (int, int, error) {
	startIdx := 0
	var err error
	if start != nil {
		startIdx, err = normalizeIndex(start, length)
		if err != nil {
			return 0, 0, err
		}
	}
	endIdx := length
	if end != nil {
		// length is a valid end slice index, but not a valid ordinary index (out of bounds)
		if endNum, ok := end.(*NumVal); ok && int(endNum.V) != length {
			endIdx, err = normalizeIndex(end, length)
			if err != nil {
				return 0, 0, err
			}
		}
	}
	if startIdx > endIdx {
		return 0, 0, fmt.Errorf("%w: %d > %d", ErrSlice, startIdx, endIdx)
	}
	return startIdx, endIdx, nil
}

func normalizeIndex(idx Value, length int) (int, error) {
	index, ok := idx.(*NumVal)
	if !ok {
		return 0, fmt.Errorf("%w: expected num, found %v", ErrType, idx.Type())
	}
	i := int(index.V)
	if i < -length || i >= length {
		return 0, fmt.Errorf("%w: %d", ErrBounds, i)
	}
	if i < 0 {
		return length + i, nil // -1 references len-1 i.e. last element
	}
	return i, nil
}

func Zero(t *parser.Type) Value {
	switch {
	case t == parser.NUM_TYPE:
		return &NumVal{}
	case t == parser.STRING_TYPE:
		return &StringVal{}
	case t == parser.BOOL_TYPE:
		return &BoolVal{}
	case t == parser.ANY_TYPE:
		return &AnyVal{V: &BoolVal{}}
	case t.Name == parser.ARRAY:
		elements := []Value{}
		return &ArrayVal{Elements: &elements, T: t}
	case t.Name == parser.MAP:
		order := []string{}
		return &MapVal{Pairs: map[string]Value{}, Order: &order, T: t}
	}
	panic("cannot create zero value for type " + t.String())
}

func ValueFromAny(t *parser.Type, v any) (Value, error) {
	switch {
	case t == parser.NUM_TYPE:
		val, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: expected number, found %v", ErrAnyConversion, val)
		}
		return &NumVal{V: val}, nil
	case t == parser.STRING_TYPE:
		val, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%w: expected string, found %v", ErrAnyConversion, val)
		}
		return &StringVal{V: val}, nil
	case t == parser.BOOL_TYPE:
		val, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: expected bool, found %v", ErrAnyConversion, val)
		}
		return &BoolVal{V: val}, nil
	}
	return nil, fmt.Errorf("%w: cannot create value for type %v", ErrAnyConversion, t)
}

func unwrapBasicvalue(val Value) any {
	switch v := val.(type) {
	case *NumVal:
		return v.V
	case *StringVal:
		return v.V
	case *BoolVal:
		return v.V
	case *AnyVal:
		return unwrapBasicvalue(v.V)
	default:
		return v.String()
	}
}
