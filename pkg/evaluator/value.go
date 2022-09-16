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

type Error struct {
	Message string
}

func (n *Num) Type() ValueType { return NUM }
func (n *Num) String() string  { return strconv.FormatFloat(n.Val, 'f', -1, 64) }

func (s *String) Type() ValueType { return STRING }
func (s *String) String() string  { return s.Val }

func (*Bool) Type() ValueType { return BOOL }
func (s *Bool) String() string {
	return strconv.FormatBool(s.Val)
}

func (r *ReturnValue) Type() ValueType { return RETURN_VALUE }
func (r *ReturnValue) String() string  { return r.Val.String() }

func (e *Error) Type() ValueType { return ERROR }
func (e *Error) String() string  { return "ERROR: " + e.Message }
func isError(val Value) bool { // TODO: replace with panic flow
	return val != nil && val.Type() == ERROR
}
func newError(msg string) *Error {
	return &Error{Message: msg}
}

func (a *Array) Type() ValueType { return ARRAY }
func (a *Array) String() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.String()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

func (m *Map) Type() ValueType { return MAP }
func (m *Map) String() string {
	pairs := make([]string, 0, len(m.Pairs))
	for key, value := range m.Pairs {
		pairs = append(pairs, key+":"+value.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}
