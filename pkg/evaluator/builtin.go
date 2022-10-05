package evaluator

import (
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/parser"
)

type Builtin struct {
	Func BuiltinFunc
	Decl *parser.FuncDecl
}

type Builtins map[string]Builtin

func (b Builtins) Decls() map[string]*parser.FuncDecl {
	decls := make(map[string]*parser.FuncDecl, len(b))
	for name, builtin := range b {
		decls[name] = builtin.Decl
	}
	return decls
}

type BuiltinFunc func(args []Value) Value

func (b BuiltinFunc) Type() ValueType { return BUILTIN }
func (b BuiltinFunc) String() string  { return "builtin function" }

func DefaultBuiltins(printFn func(string)) Builtins {
	return Builtins{
		"print": {Func: printFunc(printFn), Decl: printDecl},
		"len":   {Func: BuiltinFunc(lenFunc), Decl: lenDecl},
		"move":  {Func: moveFunc(printFn), Decl: moveDecl},
		"line":  {Func: lineFunc(printFn), Decl: lineDecl},
	}
}

var printDecl = &parser.FuncDecl{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) BuiltinFunc {
	return func(args []Value) Value {
		argList := make([]string, len(args))
		for i, arg := range args {
			argList[i] = arg.String()
		}
		printFn(strings.Join(argList, " ") + "\n")
		return nil
	}
}

var lenDecl = &parser.FuncDecl{
	Name:       "len",
	Params:     []*parser.Var{{Name: "a", T: parser.ANY_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func lenFunc(args []Value) Value {
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
	}
	return newError("'len' takes 1 argument of type 'string', array '[]' or map '{}' not " + args[0].Type().String())
}

var moveDecl = &parser.FuncDecl{
	Name: "move",
	Params: []*parser.Var{
		{Name: "x", T: parser.NUM_TYPE},
		{Name: "y", T: parser.NUM_TYPE},
	},
	ReturnType: parser.NUM_TYPE,
}

func moveFunc(printFn func(string)) BuiltinFunc {
	return func(args []Value) Value {
		printFn("'move' not yet implemented\n")
		return nil
	}
}

var lineDecl = &parser.FuncDecl{
	Name: "line",
	Params: []*parser.Var{
		{Name: "x", T: parser.NUM_TYPE},
		{Name: "y", T: parser.NUM_TYPE},
	},
	ReturnType: parser.NUM_TYPE,
}

func lineFunc(printFn func(string)) BuiltinFunc {
	return func(args []Value) Value {
		printFn("'line' not yet implemented\n")
		return nil
	}
}
