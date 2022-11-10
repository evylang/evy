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

type Builtins struct {
	Funcs map[string]Builtin
	Print func(s string)
}

func (b Builtins) Decls() map[string]*parser.FuncDecl {
	decls := make(map[string]*parser.FuncDecl, len(b.Funcs))
	for name, builtin := range b.Funcs {
		decls[name] = builtin.Decl
	}
	return decls
}

type BuiltinFunc func(args []Value) Value

func (b BuiltinFunc) Type() ValueType { return BUILTIN }
func (b BuiltinFunc) String() string  { return "builtin function" }

func DefaultBuiltins(rt Runtime) Builtins {
	funcs := map[string]Builtin{
		"print":  {Func: printFunc(rt.Print), Decl: printDecl},
		"sprint": {Func: sprintFunc, Decl: sprintDecl},
		"join":   {Func: joinFunc, Decl: joinDecl},
		"split":  {Func: splitFunc, Decl: splitDecl},

		"len": {Func: BuiltinFunc(lenFunc), Decl: lenDecl},
		"has": {Func: BuiltinFunc(hasFunc), Decl: hasDecl},
		"del": {Func: BuiltinFunc(delFunc), Decl: delDecl},

		"move":   xyBuiltin("move", rt.Graphics.Move, rt.Print),
		"line":   xyBuiltin("line", rt.Graphics.Line, rt.Print),
		"rect":   xyBuiltin("rect", rt.Graphics.Rect, rt.Print),
		"circle": numBuiltin("circle", rt.Graphics.Circle, rt.Print),
		"width":  numBuiltin("width", rt.Graphics.Width, rt.Print),
		"color":  stringBuiltin("color", rt.Graphics.Color, rt.Print),
		"colour": stringBuiltin("colour", rt.Graphics.Color, rt.Print),
	}
	return Builtins{Funcs: funcs, Print: rt.Print}
}

type Runtime struct {
	Print    func(string)
	Graphics GraphicsRuntime
}

type GraphicsRuntime struct {
	Move   func(x, y float64)
	Line   func(x, y float64)
	Rect   func(dx, dy float64)
	Circle func(radius float64)
	Width  func(w float64)
	Color  func(s string)
}

var printDecl = &parser.FuncDecl{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) BuiltinFunc {
	return func(args []Value) Value {
		printFn(join(args, " ") + "\n")
		return nil
	}
}

var sprintDecl = &parser.FuncDecl{
	Name:          "sprint",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.STRING_TYPE,
}

func sprintFunc(args []Value) Value {
	return &String{Val: join(args, " ")}
}

var joinDecl = &parser.FuncDecl{
	Name: "join",
	Params: []*parser.Var{
		{Name: "arr", T: parser.GENERIC_ARRAY},
		{Name: "sep", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func joinFunc(args []Value) Value {
	arr := args[0].(*Array)
	sep := args[1].(*String)
	s := join(*arr.Elements, sep.Val)
	return &String{Val: s}
}

func join(args []Value, sep string) string {
	argStrings := make([]string, len(args))
	for i, arg := range args {
		argStrings[i] = arg.String()
	}
	return strings.Join(argStrings, sep)
}

var stringArrayType *parser.Type = &parser.Type{
	Name: parser.ARRAY,
	Sub:  parser.STRING_TYPE,
}

var splitDecl = &parser.FuncDecl{
	Name: "split",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "sep", T: parser.STRING_TYPE},
	},
	ReturnType: stringArrayType,
}

func splitFunc(args []Value) Value {
	s := args[0].(*String)
	sep := args[1].(*String)
	slice := strings.Split(s.Val, sep.Val)
	elements := make([]Value, len(slice))
	for i, s := range slice {
		elements[i] = &String{Val: s}
	}
	return &Array{Elements: &elements}
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
		return &Num{Val: float64(len(*arg.Elements))}
	case *String:
		return &Num{Val: float64(len(arg.Val))}
	}
	return newError("'len' takes 1 argument of type 'string', array '[]' or map '{}' not " + args[0].Type().String())
}

var hasDecl = &parser.FuncDecl{
	Name: "has",
	Params: []*parser.Var{
		{Name: "m", T: parser.GENERIC_MAP},
		{Name: "key", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func hasFunc(args []Value) Value {
	m := args[0].(*Map)
	key := args[1].(*String)
	_, ok := m.Pairs[key.Val]
	return &Bool{Val: ok}
}

var delDecl = &parser.FuncDecl{
	Name: "del",
	Params: []*parser.Var{
		{Name: "m", T: parser.GENERIC_MAP},
		{Name: "key", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NONE_TYPE,
}

func delFunc(args []Value) Value {
	m := args[0].(*Map)
	keyStr := args[1].(*String)
	m.Delete(keyStr.Val)
	return nil
}

func xyDecl(name string) *parser.FuncDecl {
	return &parser.FuncDecl{
		Name: name,
		Params: []*parser.Var{
			{Name: "x", T: parser.NUM_TYPE},
			{Name: "y", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func xyBuiltin(name string, fn func(x, y float64), printFn func(string)) Builtin {
	result := Builtin{Decl: xyDecl(name)}
	if fn == nil {
		result.Func = notImplementedFunc(name, printFn)
		return result
	}
	result.Func = func(args []Value) Value {
		x := args[0].(*Num)
		y := args[1].(*Num)
		fn(x.Val, y.Val)
		return nil
	}
	return result
}

func numDecl(name string) *parser.FuncDecl {
	return &parser.FuncDecl{
		Name: name,
		Params: []*parser.Var{
			{Name: "n", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func numBuiltin(name string, fn func(n float64), printFn func(string)) Builtin {
	result := Builtin{Decl: numDecl(name)}
	if fn == nil {
		result.Func = notImplementedFunc(name, printFn)
		return result
	}
	result.Func = func(args []Value) Value {
		n := args[0].(*Num)
		fn(n.Val)
		return nil
	}
	return result
}

func stringDecl(name string) *parser.FuncDecl {
	return &parser.FuncDecl{
		Name: name,
		Params: []*parser.Var{
			{Name: "str", T: parser.STRING_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func stringBuiltin(name string, fn func(str string), printFn func(string)) Builtin {
	result := Builtin{Decl: stringDecl(name)}
	if fn == nil {
		result.Func = notImplementedFunc(name, printFn)
		return result
	}
	result.Func = func(args []Value) Value {
		str := args[0].(*String)
		fn(str.Val)
		return nil
	}
	return result
}

func notImplementedFunc(name string, printFn func(string)) BuiltinFunc {
	return func(args []Value) Value {
		printFn("'" + name + "' not yet implemented\n")
		return nil
	}
}
