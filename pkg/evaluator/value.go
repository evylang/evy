package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/parser"
)

type ValueType int

const (
	NUM ValueType = iota
	BOOL
	STRING
	ANY
	ARRAY
	MAP
	RETURN_VALUE
	BREAK
	FUNCTION
	BUILTIN
)

var valueTypeStrings = map[ValueType]string{
	NUM:          "num",
	BOOL:         "bool",
	STRING:       "string",
	ANY:          "any",
	ARRAY:        "array",
	MAP:          "map",
	RETURN_VALUE: "return_value",
	FUNCTION:     "function",
	BUILTIN:      "builtin",
}

func (t ValueType) String() string {
	if s, ok := valueTypeStrings[t]; ok {
		return s
	}
	return "<UNKNOWN>"
}

func (t ValueType) GoString() string {
	return t.String()
}

type Value interface {
	Type() ValueType
	Equals(Value) bool
	String() string
	Set(Value)
}

type Num struct {
	Val float64
}

type Bool struct {
	Val bool
}

type String struct {
	Val       string
	runeSlice []rune
}

type Any struct {
	Val Value
}

type Array struct {
	Elements *[]Value
}

type Map struct {
	Pairs map[string]Value
	Order *[]string
}

type ReturnValue struct {
	Val Value
}

type Break struct{}

func (n *Num) Type() ValueType { return NUM }
func (n *Num) String() string  { return strconv.FormatFloat(n.Val, 'f', -1, 64) }
func (n *Num) Equals(v Value) bool {
	n2, ok := v.(*Num)
	if !ok {
		panic("internal error: Num.Equals called with non-Num Value")
	}
	return n.Val == n2.Val
}

func (n *Num) Set(v Value) {
	n2, ok := v.(*Num)
	if !ok {
		panic("internal error: Num.Set called with with non-Num Value")
	}
	*n = *n2
}

func (s *String) Type() ValueType { return STRING }
func (s *String) String() string  { return s.Val }
func (s *String) Equals(v Value) bool {
	s2, ok := v.(*String)
	if !ok {
		panic("internal error: String.Equals called with non-String Value")
	}
	return s.Val == s2.Val
}

func (s *String) Set(v Value) {
	s2, ok := v.(*String)
	if !ok {
		panic("internal error: String.Set called with with non-String Value")
	}
	*s = *s2
}

func (s *String) runes() []rune {
	if s.runeSlice == nil {
		s.runeSlice = []rune(s.Val)
	}
	return s.runeSlice
}

func (s *String) Index(idx Value) (Value, error) {
	runes := s.runes()
	i, err := normalizeIndex(idx, len(runes))
	if err != nil {
		return nil, err
	}
	return &String{Val: string(runes[i])}, nil
}

func (s *String) Slice(start, end Value) (Value, error) {
	runes := s.runes()
	length := len(runes)
	startIdx, endIdx, err := normalizeSliceIndices(start, end, length)
	if err != nil {
		return nil, err
	}
	return &String{Val: string(runes[startIdx:endIdx])}, nil
}

func (*Bool) Type() ValueType { return BOOL }
func (b *Bool) String() string {
	return strconv.FormatBool(b.Val)
}

func (b *Bool) Equals(v Value) bool {
	b2, ok := v.(*Bool)
	if !ok {
		panic("internal error: Bool.Equals called with non-Bool Value")
	}
	return b.Val == b2.Val
}

func (b *Bool) Set(v Value) {
	b2, ok := v.(*Bool)
	if !ok {
		panic("internal error: Bool.Set called with with non-Bool Value")
	}
	*b = *b2
}

func (*Any) Type() ValueType { return ANY }
func (a *Any) String() string {
	return a.Val.String()
}

func (a *Any) Equals(v Value) bool {
	a2, ok := v.(*Any)
	if !ok {
		panic("internal error: Any.Equals called with non-Any Value")
	}
	return a.Val.Equals(a2.Val)
}

func (a *Any) Set(v Value) {
	if a2, ok := v.(*Any); ok {
		a.Val = copyOrRef(a2.Val)
	} else {
		a.Val = copyOrRef(v)
	}
}

func (r *ReturnValue) Type() ValueType     { return RETURN_VALUE }
func (r *ReturnValue) String() string      { return r.Val.String() }
func (r *ReturnValue) Equals(v Value) bool { return r.Val.Equals(v) }
func (r *ReturnValue) Set(v Value)         { r.Val.Set(v) }

func (r *Break) Type() ValueType     { return BREAK }
func (r *Break) String() string      { return "" }
func (r *Break) Equals(_ Value) bool { return false }
func (r *Break) Set(_ Value)         {}

func (a *Array) Type() ValueType { return ARRAY }
func (a *Array) String() string {
	elements := make([]string, len(*a.Elements))
	for i, e := range *a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, " ") + "]"
}

func (a *Array) Equals(v Value) bool {
	a2, ok := v.(*Array)
	if !ok {
		panic("internal error: Array.Equals called with non-Array Value")
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

func (a *Array) Set(v Value) {
	a2, ok := v.(*Array)
	if !ok {
		panic("internal error: Array.Set called with with non-Array Value")
	}
	*a = *a2
}

func (a *Array) Index(idx Value) (Value, error) {
	i, err := normalizeIndex(idx, len(*a.Elements))
	if err != nil {
		return nil, err
	}
	elements := *a.Elements
	return elements[i], nil
}

func (a *Array) Copy() *Array {
	elements := make([]Value, len(*a.Elements))
	for i, v := range *a.Elements {
		elements[i] = copyOrRef(v)
	}
	return &Array{Elements: &elements}
}

func (a *Array) Slice(start, end Value) (Value, error) {
	length := len(*a.Elements)
	startIdx, endIdx, err := normalizeSliceIndices(start, end, length)
	if err != nil {
		return nil, err
	}

	elements := make([]Value, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		v := (*a.Elements)[i]
		elements[i-startIdx] = copyOrRef(v)
	}
	return &Array{Elements: &elements}, nil
}

// copyOrRef is a copy of the input value for basic types and a
// reference to the value for composite types (arrays and maps).
func copyOrRef(val Value) Value {
	switch v := val.(type) {
	case *Num:
		return &Num{Val: v.Val}
	case *String:
		return &String{Val: v.Val}
	case *Bool:
		return &Bool{Val: v.Val}
	case *Any:
		return &Any{Val: copyOrRef(v.Val)}
	case *Array:
		return v
	case *Map:
		return v
	}
	panic("internal error: copyOrRef called with with invalid Value")
}

func (m *Map) Type() ValueType { return MAP }
func (m *Map) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for _, key := range *m.Order {
		pairs = append(pairs, key+":"+m.Pairs[key].String())
	}
	return "{" + strings.Join(pairs, " ") + "}"
}

func (m *Map) Equals(v Value) bool {
	m2, ok := v.(*Map)
	if !ok {
		panic("internal error: Map.Equals called with non-Map Value")
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

func (m *Map) Set(v Value) {
	m2, ok := v.(*Map)
	if !ok {
		panic("internal error: Map.Set called with with non-Map Value")
	}
	*m = *m2
}

func (m *Map) Get(key string) (Value, error) {
	val, ok := m.Pairs[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrMapKey, key)
	}
	return val, nil
}

func (m *Map) InsertKey(key string, t *parser.Type) {
	if _, ok := m.Pairs[key]; ok {
		return
	}
	*m.Order = append(*m.Order, key)
	m.Pairs[key] = zero(t)
}

func (m *Map) Delete(key string) {
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

func isReturn(val Value) bool {
	return val != nil && val.Type() == RETURN_VALUE
}

func isBreak(val Value) bool {
	return val != nil && val.Type() == BREAK
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
		if endNum, ok := end.(*Num); ok && int(endNum.Val) != length {
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
	index, ok := idx.(*Num)
	if !ok {
		return 0, fmt.Errorf("%w: expected num, found %v", ErrType, idx.Type())
	}
	i := int(index.Val)
	if i < -length || i >= length {
		return 0, fmt.Errorf("%w: %d", ErrBounds, i)
	}
	if i < 0 {
		return length + i, nil // -1 references len-1 i.e. last element
	}
	return i, nil
}

func zero(t *parser.Type) Value {
	switch {
	case t == parser.NUM_TYPE:
		return &Num{}
	case t == parser.STRING_TYPE:
		return &String{}
	case t == parser.BOOL_TYPE:
		return &Bool{}
	case t == parser.ANY_TYPE:
		return &Any{Val: &Bool{}}
	case t.Name == parser.ARRAY:
		elements := []Value{}
		return &Array{Elements: &elements}
	case t.Name == parser.MAP:
		order := []string{}
		return &Map{Pairs: map[string]Value{}, Order: &order}
	}
	panic("cannot create zero value for type " + t.String())
}

func valueFromAny(t *parser.Type, v any) (Value, error) {
	switch {
	case t == parser.NUM_TYPE:
		val, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: expected number, found %v", ErrAnyConversion, val)
		}
		return &Num{Val: val}, nil
	case t == parser.STRING_TYPE:
		val, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%w: expected string, found %v", ErrAnyConversion, val)
		}
		return &String{Val: val}, nil
	case t == parser.BOOL_TYPE:
		val, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: expected bool, found %v", ErrAnyConversion, val)
		}
		return &Bool{Val: val}, nil
	}
	return nil, fmt.Errorf("%w: cannot create value for type %v", ErrAnyConversion, t)
}

func unwrapBasicValue(val Value) any {
	switch v := val.(type) {
	case *Num:
		return v.Val
	case *String:
		return v.Val
	case *Bool:
		return v.Val
	case *Any:
		return unwrapBasicValue(v.Val)
	default:
		return v.String()
	}
}
