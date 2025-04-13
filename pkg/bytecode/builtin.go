package bytecode

import (
	"strings"

	"evylang.dev/evy/pkg/parser"
)

type builtin struct {
	Name string
	Func builtinFunc
	Decl *parser.FuncDefStmt
}

type builtins struct {
	Funcs    []builtin
	Platform Platform
}

func newBuiltins(rt Platform) builtins {
	return builtins{
		Funcs: []builtin{
			{Name: "join", Func: joinFunc, Decl: joinDecl},
			{Name: "print", Func: printFunc(rt.Print), Decl: printDecl},
		},
	}
}

// BuiltinDecls returns the signatures of all built-in functions and
// event handlers, as well as predefined global variables, for use by
// the [parser.Parse] function.
func BuiltinDecls(platform Platform) parser.Builtins {
	b := newBuiltins(platform)
	return builtinsDecls(b)
}

func builtinsDecls(b builtins) parser.Builtins {
	funcs := make(map[string]*parser.FuncDefStmt, len(b.Funcs))
	for _, builtin := range b.Funcs {
		funcs[builtin.Name] = builtin.Decl
	}
	return parser.Builtins{
		Funcs: funcs,
	}
}

type builtinFunc func(args ...value) (value, error)

func (builtinFunc) Type() *parser.Type {
	return parser.NONE_TYPE
}

func (f builtinFunc) String() string {
	return "builtin"
}

func (f builtinFunc) Equals(v value) bool {
	_, ok := v.(builtinFunc)
	if !ok {
		panic("internal error: builtin.Equals called with non-builtin value")
	}
	return true
}

var printDecl = &parser.FuncDefStmt{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) builtinFunc {
	return func(args ...value) (value, error) {
		// TODO: Having to do this feels a bit suspicious
		printFn(join(args[0].(arrayVal).Elements, " ") + "\n")
		return &noneVal{}, nil
	}
}

var joinDecl = &parser.FuncDefStmt{
	Name: "join",
	Params: []*parser.Var{
		{Name: "arr", T: parser.GENERIC_ARRAY},
		{Name: "sep", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func joinFunc(args ...value) (value, error) {
	arr := args[0].(arrayVal)
	sep := args[1].(stringVal)
	s := join(arr.Elements, sep)
	return stringVal(s), nil
}

func join(args []value, sep stringVal) string {
	argStrings := make([]string, len(args))
	for i, arg := range args {
		argStrings[i] = arg.String()
	}
	return strings.Join(argStrings, string(sep))
}
