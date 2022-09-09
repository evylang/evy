package evaluator

import (
	"strconv"
	"strings"
)

type Builtin func(args []Value) Value

func (b Builtin) Type() ValueType { return BUILTIN }
func (b Builtin) String() string  { return "builtin function" }

func newBuiltins(e *Evaluator) map[string]Builtin {
	return map[string]Builtin{
		"print": Builtin(e.Print),
		"len":   Builtin(Len),
	}
}

func (e *Evaluator) Print(args []Value) Value {
	argList := make([]string, len(args))
	for i, arg := range args {
		argList[i] = arg.String()
	}
	e.print(strings.Join(argList, " "))
	return nil
}

func Len(args []Value) Value {
	if len(args) != 1 {
		return newError("'len' takes 1 argument not " + strconv.Itoa(len(args)))
	}
	switch arg := args[0].(type) {
	case *Map:
		return &Num{Val: float64(len(arg.Pairs))}
	case *Array:
		return &Num{Val: float64(len(arg.Elements))}
	case *String:
		return &Num{Val: float64(len(arg.Val))}
	default:
		return newError("'len' takes 1 argument of type 'string', array '[]' or map '{}' not " + args[0].Type().String())
	}

}
