package evaluator

import (
	"strconv"
	"strings"
)

type ValueType int

const (
	ERROR ValueType = iota
	NUM
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
	ERROR:        "ERROR",
	NUM:          "NUM",
	BOOL:         "BOOL",
	STRING:       "STRING",
	ANY:          "ANY",
	ARRAY:        "ARRAY",
	MAP:          "MAP",
	RETURN_VALUE: "RETURN_VALUE",
	FUNCTION:     "FUNCTION",
	BUILTIN:      "BUILTIN",
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
	Equals(Value) bool // TODO: panic if wrong type
	String() string    // TODO: panic if wrong type
	Set(Value)
}

type Num struct {
	Val float64
}

type Bool struct {
	Val bool
}

type String struct {
	Val   string
	runes []rune
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

type Error struct {
	Message string
}

func (n *Num) Type() ValueType { return NUM }
func (n *Num) String() string  { return strconv.FormatFloat(n.Val, 'f', -1, 64) }
func (n *Num) Equals(v Value) bool {
	if n2, ok := v.(*Num); ok {
		return n.Val == n2.Val
	}
	return false // TODO: panic here when reworking ErrValue to panics; same in all Equals methods
}

func (n *Num) Set(v Value) {
	if n2, ok := v.(*Num); ok {
		*n = *n2
	}
	// TODO: panic here when reworking ErrValue to panics; same in all Set methods
}

func (s *String) Type() ValueType { return STRING }
func (s *String) String() string  { return s.Val }
func (s *String) Equals(v Value) bool {
	if s2, ok := v.(*String); ok {
		return s.Val == s2.Val
	}
	return false
}

func (s *String) Set(v Value) {
	if s2, ok := v.(*String); ok {
		*s = *s2
	}
}

func (s *String) Index(idx Value) Value {
	if s.runes == nil {
		s.runes = []rune(s.Val)
	}
	i, err := normalizeIndex(idx, len(s.runes))
	if err != nil {
		return err
	}
	return &String{Val: string(s.runes[i])}
}

func (*Bool) Type() ValueType { return BOOL }
func (b *Bool) String() string {
	return strconv.FormatBool(b.Val)
}

func (b *Bool) Equals(v Value) bool {
	if b2, ok := v.(*Bool); ok {
		return b.Val == b2.Val
	}
	return false
}

func (b *Bool) Set(v Value) {
	if b2, ok := v.(*Bool); ok {
		*b = *b2
	}
}

func (*Any) Type() ValueType { return ANY }
func (a *Any) String() string {
	return a.Val.String()
}

func (a *Any) Equals(v Value) bool {
	return a.Val.Equals(v)
}

func (a *Any) Set(v Value) {
	a.Val = v
}

func (r *ReturnValue) Type() ValueType     { return RETURN_VALUE }
func (r *ReturnValue) String() string      { return r.Val.String() }
func (r *ReturnValue) Equals(v Value) bool { return r.Val.Equals(v) }
func (r *ReturnValue) Set(v Value)         { r.Val.Set(v) }

func (r *Break) Type() ValueType     { return BREAK }
func (r *Break) String() string      { return "" }
func (r *Break) Equals(_ Value) bool { return false }
func (r *Break) Set(_ Value)         {}

func (e *Error) Type() ValueType     { return ERROR }
func (e *Error) String() string      { return "ERROR: " + e.Message }
func (e *Error) Equals(_ Value) bool { return false }
func (e *Error) Set(_ Value)         {}

func (a *Array) Type() ValueType { return ARRAY }
func (a *Array) String() string {
	elements := make([]string, len(*a.Elements))
	for i, e := range *a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, " ") + "]"
}

func (a *Array) Equals(v Value) bool {
	if a2, ok := v.(*Array); ok {
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
	return false
}

func (a *Array) Set(v Value) {
	if a2, ok := v.(*Array); ok {
		*a = *a2
	}
}

func (a *Array) Index(idx Value) Value {
	i, err := normalizeIndex(idx, len(*a.Elements))
	if err != nil {
		return err
	}
	elements := *a.Elements
	return elements[i]
}

func (a *Array) Copy() *Array {
	elements := make([]Value, len(*a.Elements))
	for i, v := range *a.Elements {
		elements[i] = passedVal(v)
	}
	return &Array{Elements: &elements}
}

// passedVal is a pass by reference or copy of the value depending on type.
func passedVal(val Value) Value {
	switch v := val.(type) {
	case *Num:
		return &Num{Val: v.Val}
	case *String:
		return &String{Val: v.Val}
	case *Bool:
		return &Bool{Val: v.Val}
	case *Any:
		return &Any{Val: passedVal(v.Val)}
	case *Array:
		return v
	case *Map:
		return v
	}
	return nil // TODO: panic
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
	if m2, ok := v.(*Map); ok {
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
	return false
}

func (m *Map) Set(v Value) {
	if m2, ok := v.(*Map); ok {
		*m = *m2
	}
}

func (m *Map) Get(key string) Value {
	val, ok := m.Pairs[key]
	if !ok {
		return newError("no value for key " + key)
	}
	return val
}

func isError(val Value) bool { // TODO: replace with panic flow
	return val != nil && val.Type() == ERROR
}

func isReturn(val Value) bool {
	return val != nil && val.Type() == RETURN_VALUE
}

func isBreak(val Value) bool {
	return val != nil && val.Type() == BREAK
}

func newError(msg string) *Error {
	return &Error{Message: msg}
}

func normalizeIndex(idx Value, length int) (int, Value) {
	index, ok := idx.(*Num)
	if !ok {
		return 0, newError("expected index of type num, found " + idx.Type().String())
	}
	i := int(index.Val)
	if i < -length || i >= length {
		boundsStr := strconv.Itoa(-length) + " and " + strconv.Itoa(length-1)
		msg := "index " + strconv.Itoa(i) + " out of bounds, should be between " + boundsStr
		return 0, newError(msg)
	}
	if i < 0 {
		return length + i, nil // -1 references len-1 i.e. last element
	}
	return i, nil
}
