package bytecode

import (
	"strconv"

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

func (n *numVal) Type() *parser.Type { return parser.NUM_TYPE }

func (n *numVal) String() string { return strconv.FormatFloat(n.V, 'f', -1, 64) }

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
