package evaluator

import (
	"strconv"
	"strings"
)

var (
	TRUE  = &Bool{Value: true}
	FALSE = &Bool{Value: false}
)

func boolObject(b bool) *Bool {
	if b {
		return TRUE
	}
	return FALSE
}

type ObjectType int

const (
	ERROR ObjectType = iota
	NUM
	BOOL
	STRING
	ARRAY
	MAP
	RETURN_VALUE
	FUNCTION
	BUILTIN
)

var objectTypeStrings = map[ObjectType]string{
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

func (t ObjectType) String() string {
	if s, ok := objectTypeStrings[t]; ok {
		return s
	}
	return "<UNKNOWN>"
}

func (t ObjectType) GoString() string {
	return t.String()
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Num struct {
	Value float64
}

type Bool struct {
	Value bool
}

type String struct {
	Value string
}

type Array struct {
	Elements []Object
}

type Map struct {
	Pairs map[string]Object
}

type ReturnValue struct {
	Value Object
}

type Error struct {
	Message string
}

func (n *Num) Type() ObjectType { return NUM }
func (n *Num) Inspect() string  { return strconv.FormatFloat(n.Value, 'f', -1, 64) }

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }

func (*Bool) Type() ObjectType { return BOOL }
func (s *Bool) Inspect() string {
	if s.Value {
		return "true"
	}
	return "false"
}

func (r *ReturnValue) Type() ObjectType { return RETURN_VALUE }
func (r *ReturnValue) Inspect() string  { return r.Value.Inspect() }

func (e *Error) Type() ObjectType { return ERROR }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func isError(obj Object) bool {
	return obj != nil && obj.Type() == ERROR
}
func newError(msg string) *Error {
	return &Error{Message: msg}
}

func (a *Array) Type() ObjectType { return ARRAY }
func (a *Array) Inspect() string {
	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.Inspect()
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

func (m *Map) Type() ObjectType { return MAP }
func (m *Map) Inspect() string {
	pairs := make([]string, 0, len(m.Pairs))
	for key, value := range m.Pairs {
		pairs = append(pairs, key+":"+value.Inspect())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}
