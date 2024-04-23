package bytecode

import (
	"fmt"
	"strconv"
	"strings"

	"evylang.dev/evy/pkg/parser"
)

var (
	// ErrBounds reports an index out of bounds in an array or string.
	ErrBounds = fmt.Errorf("%w: index out of bounds", ErrPanic)
	// ErrIndexValue reports an index value error if the index is not an integer, e.g. 1.1.
	ErrIndexValue = fmt.Errorf("%w: index not an integer", ErrPanic)
	// ErrMapKey reports that no value was found for a specific key
	// used in a map index.
	ErrMapKey = fmt.Errorf("%w: no value for map key", ErrPanic)
	// ErrSlice reports an invalid slice where start index > end index.
	ErrSlice = fmt.Errorf("%w: invalid slice", ErrPanic)
)

type indexType int

const (
	indexExpression indexType = iota
	sliceExpression
)

type value interface {
	Type() *parser.Type
	Equals(value) bool
	String() string
}

type indexable interface {
	// Index returns the value at the specified key. A user error will
	// be returned if the key is not found inside the structure.
	Index(key value) (value, error)
}

type sliceable interface {
	// Slice returns a subset of elements between the start (inclusive)
	// and end (exclusive) index. A user error will be returned if start
	// or end index are outside the bounds of the structure.
	Slice(start, end value) (value, error)
}

type numVal float64

func (n numVal) Type() *parser.Type { return parser.NUM_TYPE }

func (n numVal) String() string {
	return strconv.FormatFloat(float64(n), 'f', -1, 64)
}

func (n numVal) Equals(v value) bool {
	n2, ok := v.(numVal)
	if !ok {
		panic("internal error: Num.Equals called with non-Num value")
	}
	return n == n2
}

type boolVal bool

func (boolVal) Type() *parser.Type { return parser.BOOL_TYPE }

func (b boolVal) String() string {
	return strconv.FormatBool(bool(b))
}

func (b boolVal) Equals(v value) bool {
	b2, ok := v.(boolVal)
	if !ok {
		panic("internal error: Bool.Equals called with non-Bool value")
	}
	return b == b2
}

type stringVal string

func (s stringVal) Type() *parser.Type { return parser.STRING_TYPE }

func (s stringVal) String() string {
	return string(s)
}

func (s stringVal) Equals(v value) bool {
	s2, ok := v.(stringVal)
	if !ok {
		panic("internal error: String.Equals called with non-String value")
	}
	return s == s2
}

func (s stringVal) Index(idx value) (value, error) {
	index, err := normalizeIndex(idx, len(s), indexExpression)
	if err != nil {
		return nil, err
	}
	return s[index : index+1], nil
}

func (s stringVal) Slice(start, end value) (value, error) {
	startIdx, endIdx, err := normalizeSliceIndices(start, end, len(s))
	if err != nil {
		return nil, err
	}
	return s[startIdx:endIdx], nil
}

type arrayVal struct {
	Elements []value
}

func (a arrayVal) Type() *parser.Type {
	// Revisit this when adding the typeof builtin: https://github.com/evylang/evy/pull/305#discussion_r1531149977
	return parser.GENERIC_ARRAY
}

func (a arrayVal) String() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, " ") + "]"
}

func (a arrayVal) Equals(v value) bool {
	a2, ok := v.(arrayVal)
	if !ok {
		panic("internal error: Array.Equals called with non-Array value")
	}
	if len(a.Elements) != len(a2.Elements) {
		return false
	}
	elements2 := a2.Elements
	for i, e := range a.Elements {
		e2 := elements2[i]
		if !e.Equals(e2) {
			return false
		}
	}
	return true
}

func (a arrayVal) Index(idx value) (value, error) {
	index, err := normalizeIndex(idx, len(a.Elements), indexExpression)
	if err != nil {
		return nil, err
	}
	return a.Elements[index], nil
}

// Set checks that idx is in bounds and then sets
// the value of the element at idx to val.
func (a arrayVal) Set(idx, val value) error {
	index := int(idx.(numVal))
	length := len(a.Elements)
	if index >= length || index < -length {
		return fmt.Errorf("%w: %d", ErrBounds, index)
	}
	if index < 0 {
		index += length
	}
	a.Elements[index] = val
	return nil
}

func (a arrayVal) Slice(start, end value) (value, error) {
	startIdx, endIdx, err := normalizeSliceIndices(start, end, len(a.Elements))
	if err != nil {
		return nil, err
	}
	elements := make([]value, endIdx-startIdx)
	copy(elements, a.Elements[startIdx:endIdx])
	return arrayVal{Elements: elements}, nil
}

type mapVal struct {
	order []stringVal
	m     map[stringVal]value
}

func (m mapVal) Type() *parser.Type {
	return parser.GENERIC_MAP
}

func (m mapVal) String() string {
	pairs := make([]string, len(m.order))
	for i, k := range m.order {
		pairs[i] = fmt.Sprintf("%s: %s", k, m.m[k])
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (m mapVal) Equals(v value) bool {
	m2, ok := v.(mapVal)
	if !ok {
		panic("internal error: Map.Equals called with non-Map value")
	}
	if len(m.m) != len(m2.m) {
		return false
	}
	for key, val := range m.m {
		val2, ok := m2.m[key]
		if !ok || val2 == nil || !val.Equals(val2) {
			return false
		}
	}
	return true
}

func (m mapVal) Index(idx value) (value, error) {
	k := idx.(stringVal)
	val, ok := m.m[k]
	if !ok {
		return nil, fmt.Errorf("%w %q", ErrMapKey, idx)
	}
	return val, nil
}

type noneVal struct{}

func (noneVal) Type() *parser.Type {
	return parser.NONE_TYPE
}

func (noneVal) String() string {
	return ""
}

func (noneVal) Equals(_ value) bool {
	return false
}

type funcVal struct {
	Instructions Instructions
}

func (funcVal) Type() *parser.Type {
	return parser.NONE_TYPE
}

func (f funcVal) String() string {
	return fmt.Sprintf("func[%v]", f.Instructions)
}

func (f funcVal) Equals(v value) bool {
	f2, ok := v.(funcVal)
	if !ok {
		panic("internal error: function.Equals called with non-function value")
	}
	ins2 := f2.Instructions
	if len(f.Instructions) != len(ins2) {
		return false
	}
	for i, b := range f.Instructions {
		b2 := ins2[i]
		if b != b2 {
			return false
		}
	}
	return true
}

func normalizeIndex(idx value, length int, indexType indexType) (int, error) {
	limit := length - 1
	if indexType == sliceExpression {
		limit++ // slice expression indices can index one past the end
	}

	index := idx.(numVal)
	i := int(index)

	if float64(index) != float64(i) {
		return 0, fmt.Errorf("%w: %v", ErrIndexValue, index)
	}
	if i < -length || i > limit {
		return 0, fmt.Errorf("%w: %d", ErrBounds, i)
	}
	if i < 0 {
		return length + i, nil // -1 references len-1 i.e. last element
	}
	return i, nil
}

func normalizeSliceIndices(start, end value, length int) (int, int, error) {
	startIdx := 0
	var err error
	if _, ok := start.(noneVal); !ok {
		startIdx, err = normalizeIndex(start, length, sliceExpression)
		if err != nil {
			return 0, 0, err
		}
	}
	endIdx := length
	if _, ok := end.(noneVal); !ok {
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
