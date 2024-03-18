package evaluator

import (
	"fmt"
	"strconv"
	"strings"

	"evylang.dev/evy/pkg/parser"
)

type value interface {
	Type() *parser.Type
	Equals(value) bool
	String() string
	Set(value)
}

type numVal struct {
	V float64
}

type boolVal struct {
	V bool
}

type stringVal struct {
	V         string
	runeSlice []rune
}

type anyVal struct {
	V value
}

type arrayVal struct {
	Elements *[]value
	T        *parser.Type
}

type mapVal struct {
	Pairs map[string]value
	Order *[]string
	T     *parser.Type
}

type returnVal struct {
	V value
}

type breakVal struct{}

type noneVal struct{}

func (n *numVal) Type() *parser.Type { return parser.NUM_TYPE }
func (n *numVal) String() string     { return strconv.FormatFloat(n.V, 'f', -1, 64) }
func (n *numVal) Equals(v value) bool {
	n2, ok := v.(*numVal)
	if !ok {
		panic("internal error: Num.Equals called with non-Num value")
	}
	return n.V == n2.V
}

func (n *numVal) Set(v value) {
	n2, ok := v.(*numVal)
	if !ok {
		panic("internal error: Num.Set called with with non-Num value")
	}
	*n = *n2
}

func (s *stringVal) Type() *parser.Type { return parser.STRING_TYPE }
func (s *stringVal) String() string     { return s.V }
func (s *stringVal) Equals(v value) bool {
	s2, ok := v.(*stringVal)
	if !ok {
		panic("internal error: String.Equals called with non-String value")
	}
	return s.V == s2.V
}

func (s *stringVal) Set(v value) {
	s2, ok := v.(*stringVal)
	if !ok {
		panic("internal error: String.Set called with with non-String value")
	}
	*s = *s2
}

func (s *stringVal) runes() []rune {
	if s.runeSlice == nil {
		s.runeSlice = []rune(s.V)
	}
	return s.runeSlice
}

func (s *stringVal) Index(idx value) (value, error) {
	runes := s.runes()
	i, err := normalizeIndex(idx, len(runes), indexExpression)
	if err != nil {
		return nil, err
	}
	return &stringVal{V: string(runes[i])}, nil
}

func (s *stringVal) Slice(start, end value) (value, error) {
	runes := s.runes()
	length := len(runes)
	startIdx, endIdx, err := normalizeSliceIndices(start, end, length)
	if err != nil {
		return nil, err
	}
	return &stringVal{V: string(runes[startIdx:endIdx])}, nil
}

func (*boolVal) Type() *parser.Type { return parser.BOOL_TYPE }
func (b *boolVal) String() string {
	return strconv.FormatBool(b.V)
}

func (b *boolVal) Equals(v value) bool {
	b2, ok := v.(*boolVal)
	if !ok {
		panic("internal error: Bool.Equals called with non-Bool value")
	}
	return b.V == b2.V
}

func (b *boolVal) Set(v value) {
	b2, ok := v.(*boolVal)
	if !ok {
		panic("internal error: Bool.Set called with with non-Bool value")
	}
	*b = *b2
}

func (*anyVal) Type() *parser.Type { return parser.ANY_TYPE }
func (a *anyVal) String() string {
	return a.V.String()
}

func (a *anyVal) Equals(v value) bool {
	a2, ok := v.(*anyVal)
	if !ok {
		panic("internal error: Any.Equals called with non-Any value")
	}
	return a.V.Type().Equals(a2.V.Type()) && a.V.Equals(a2.V)
}

func (a *anyVal) Set(v value) {
	if a2, ok := v.(*anyVal); ok {
		a.V = copyOrRef(a2.V)
	} else {
		a.V = copyOrRef(v)
	}
}

func (n *noneVal) Type() *parser.Type  { return parser.NONE_TYPE }
func (n *noneVal) String() string      { return "" }
func (n *noneVal) Equals(_ value) bool { return false }
func (n *noneVal) Set(_ value)         { panic("internal error: None.Set called") }

func (r *returnVal) Type() *parser.Type  { return r.V.Type() }
func (r *returnVal) String() string      { return r.V.String() }
func (r *returnVal) Equals(v value) bool { return r.V.Equals(v) }
func (r *returnVal) Set(v value)         { r.V.Set(v) }

func (r *breakVal) Type() *parser.Type  { return parser.NONE_TYPE }
func (r *breakVal) String() string      { return "" }
func (r *breakVal) Equals(_ value) bool { return false }
func (r *breakVal) Set(_ value)         {}

func (a *arrayVal) Type() *parser.Type { return a.T }
func (a *arrayVal) String() string {
	elements := make([]string, len(*a.Elements))
	for i, e := range *a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, " ") + "]"
}

func (a *arrayVal) Equals(v value) bool {
	a2, ok := v.(*arrayVal)
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

func (a *arrayVal) Set(v value) {
	a2, ok := v.(*arrayVal)
	if !ok {
		panic("internal error: Array.Set called with with non-Array value")
	}
	a.Elements = a2.Elements
}

func (a *arrayVal) Index(idx value) (value, error) {
	i, err := normalizeIndex(idx, len(*a.Elements), indexExpression)
	if err != nil {
		return nil, err
	}
	elements := *a.Elements
	return elements[i], nil
}

func (a *arrayVal) Copy() *arrayVal {
	elements := make([]value, len(*a.Elements))
	for i, v := range *a.Elements {
		elements[i] = copyOrRef(v)
	}
	return &arrayVal{Elements: &elements, T: a.T}
}

func (a *arrayVal) Slice(start, end value) (value, error) {
	length := len(*a.Elements)
	startIdx, endIdx, err := normalizeSliceIndices(start, end, length)
	if err != nil {
		return nil, err
	}

	elements := make([]value, endIdx-startIdx)
	for i := startIdx; i < endIdx; i++ {
		v := (*a.Elements)[i]
		elements[i-startIdx] = copyOrRef(v)
	}
	return &arrayVal{Elements: &elements, T: a.T}, nil
}

// copyOrRef is a copy of the input value for basic types and a
// reference to the value for composite types (arrays and maps).
func copyOrRef(val value) value {
	switch v := val.(type) {
	case *numVal:
		return &numVal{V: v.V}
	case *stringVal:
		return &stringVal{V: v.V}
	case *boolVal:
		return &boolVal{V: v.V}
	case *anyVal:
		return &anyVal{V: copyOrRef(v.V)}
	case *arrayVal:
		return v
	case *mapVal:
		return v
	}
	panic("internal error: copyOrRef called with with invalid value")
}

func (m *mapVal) Type() *parser.Type { return m.T }
func (m *mapVal) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for _, key := range *m.Order {
		pairs = append(pairs, key+":"+m.Pairs[key].String())
	}
	return "{" + strings.Join(pairs, " ") + "}"
}

func (m *mapVal) Equals(v value) bool {
	m2, ok := v.(*mapVal)
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

func (m *mapVal) Set(v value) {
	m2, ok := v.(*mapVal)
	if !ok {
		panic("internal error: Map.Set called with with non-Map value")
	}
	m.Pairs = m2.Pairs
	m.Order = m2.Order
}

func (m *mapVal) Get(key string) (value, error) {
	val, ok := m.Pairs[key]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrMapKey, key)
	}
	return val, nil
}

func (m *mapVal) InsertKey(key string, t *parser.Type) {
	if _, ok := m.Pairs[key]; ok {
		return
	}
	*m.Order = append(*m.Order, key)
	m.Pairs[key] = zero(t)
}

func (m *mapVal) Delete(key string) {
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

func isReturn(val value) bool {
	_, ok := val.(*returnVal)
	return ok
}

func isBreak(val value) bool {
	_, ok := val.(*breakVal)
	return ok
}

func normalizeSliceIndices(start, end value, length int) (int, int, error) {
	startIdx := 0
	var err error
	if start != nil {
		startIdx, err = normalizeIndex(start, length, sliceExpression)
		if err != nil {
			return 0, 0, err
		}
	}
	endIdx := length
	if end != nil {
		endIdx, err = normalizeIndex(end, length, sliceExpression)
		if err != nil {
			return 0, 0, err
		}
	}

	if startIdx > endIdx {
		return 0, 0, fmt.Errorf("%w: %d > %d", ErrSlice, startIdx, endIdx)
	}
	return startIdx, endIdx, nil
}

type indexType int

const (
	indexExpression indexType = iota
	sliceExpression
)

func normalizeIndex(idx value, length int, indexType indexType) (int, error) {
	limit := length - 1
	if indexType == sliceExpression {
		limit++ // slice expression indices can index one past the end
	}

	index, ok := idx.(*numVal)
	if !ok {
		return 0, fmt.Errorf("%w: expected num, found %v", ErrType, idx.Type())
	}
	i := int(index.V)
	if i < -length || i > limit {
		return 0, fmt.Errorf("%w: %d", ErrBounds, i)
	}
	if i < 0 {
		return length + i, nil // -1 references len-1 i.e. last element
	}
	return i, nil
}

func zero(t *parser.Type) value {
	switch {
	case t == parser.NUM_TYPE:
		return &numVal{}
	case t == parser.STRING_TYPE:
		return &stringVal{}
	case t == parser.BOOL_TYPE:
		return &boolVal{}
	case t == parser.ANY_TYPE:
		return &anyVal{V: &boolVal{}}
	case t.Name == parser.ARRAY:
		elements := []value{}
		return &arrayVal{Elements: &elements, T: t}
	case t.Name == parser.MAP:
		order := []string{}
		return &mapVal{Pairs: map[string]value{}, Order: &order, T: t}
	}
	panic("cannot create zero value for type " + t.String())
}

func valueFromAny(t *parser.Type, v any) (value, error) {
	switch {
	case t == parser.NUM_TYPE:
		val, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("%w: expected number, found %v", ErrAnyConversion, val)
		}
		return &numVal{V: val}, nil
	case t == parser.STRING_TYPE:
		val, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%w: expected string, found %v", ErrAnyConversion, val)
		}
		return &stringVal{V: val}, nil
	case t == parser.BOOL_TYPE:
		val, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("%w: expected bool, found %v", ErrAnyConversion, val)
		}
		return &boolVal{V: val}, nil
	}
	return nil, fmt.Errorf("%w: cannot create value for type %v", ErrAnyConversion, t)
}

func unwrapBasicvalue(val value) any {
	switch v := val.(type) {
	case *numVal:
		return v.V
	case *stringVal:
		return v.V
	case *boolVal:
		return v.V
	case *anyVal:
		return unwrapBasicvalue(v.V)
	default:
		return v.String()
	}
}
