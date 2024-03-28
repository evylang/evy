package bytecode

import (
	"strconv"
	"strings"

	"evylang.dev/evy/pkg/parser"
)

type value interface {
	Type() *parser.Type
	Equals(value) bool
	String() string
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
