package bytecode

import (
	"strconv"

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
