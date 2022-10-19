package evaluator

import (
	"strconv"
	"strings"
)

type ValueType int

var (
	TRUE  Value = &Bool{Val: true}
	FALSE Value = &Bool{Val: false}
)

const (
	ERROR ValueType = iota
	NUM
	BOOL
	STRING
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
	Equals(Value) bool
	String() string
}

type Num struct {
	Val float64
}

type Bool struct {
	Val bool
}

type String struct {
	Val string
}

type Array struct {
	Elements []Value
}

type Map struct {
	Pairs map[string]Value
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
	return false
}

func (s *String) Type() ValueType { return STRING }
func (s *String) String() string  { return s.Val }
func (s *String) Equals(v Value) bool {
	if s2, ok := v.(*String); ok {
		return s.Val == s2.Val
	}
	return false
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

func (r *ReturnValue) Type() ValueType     { return RETURN_VALUE }
func (r *ReturnValue) String() string      { return r.Val.String() }
func (r *ReturnValue) Equals(v Value) bool { return r.Val.Equals(v) }

func (r *Break) Type() ValueType     { return BREAK }
func (r *Break) String() string      { return "" }
func (r *Break) Equals(_ Value) bool { return false }

func (e *Error) Type() ValueType     { return ERROR }
func (e *Error) String() string      { return "ERROR: " + e.Message }
func (e *Error) Equals(_ Value) bool { return false }

func (a *Array) Type() ValueType { return ARRAY }
func (a *Array) String() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

func (a *Array) Equals(v Value) bool {
	if a2, ok := v.(*Array); ok {
		if len(a.Elements) != len(a2.Elements) {
			return false
		}
		for i, e := range a.Elements {
			e2 := a2.Elements[i]
			if !e.Equals(e2) {
				return false
			}
		}
		return true
	}
	return false
}

func (m *Map) Type() ValueType { return MAP }
func (m *Map) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for key, value := range m.Pairs {
		pairs = append(pairs, key+":"+value.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
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

func boolVal(b bool) Value {
	if b {
		return TRUE
	}
	return FALSE
}
