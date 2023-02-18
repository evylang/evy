package evaluator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"foxygo.at/evy/pkg/parser"
)

type Builtin struct {
	Func BuiltinFunc
	Decl *parser.FuncDecl
}

type Builtins struct {
	Funcs         map[string]Builtin
	Print         func(s string)
	EventHandlers map[string]*parser.EventHandler
	Globals       map[string]*parser.Var
}

func newParserBuiltins(builtins Builtins) parser.Builtins {
	funcs := make(map[string]*parser.FuncDecl, len(builtins.Funcs))
	for name, builtin := range builtins.Funcs {
		funcs[name] = builtin.Decl
	}
	return parser.Builtins{
		Funcs:         funcs,
		EventHandlers: builtins.EventHandlers,
		Globals:       builtins.Globals,
	}
}

type BuiltinFunc func(scope *scope, args []Value) Value

func (b BuiltinFunc) Type() ValueType { return BUILTIN }
func (b BuiltinFunc) String() string  { return "builtin function" }

func DefaultBuiltins(rt *Runtime) Builtins {
	funcs := map[string]Builtin{
		"read":   {Func: readFunc(rt.Read), Decl: readDecl},
		"print":  {Func: printFunc(rt.Print), Decl: printDecl},
		"printf": {Func: printfFunc(rt.Print), Decl: printDecl},

		"sprint":     {Func: sprintFunc, Decl: sprintDecl},
		"sprintf":    {Func: sprintfFunc, Decl: sprintDecl},
		"join":       {Func: joinFunc, Decl: joinDecl},
		"split":      {Func: splitFunc, Decl: splitDecl},
		"upper":      {Func: upperFunc, Decl: upperDecl},
		"lower":      {Func: lowerFunc, Decl: lowerDecl},
		"index":      {Func: indexFunc, Decl: indexDecl},
		"startswith": {Func: startswithFunc, Decl: startswithDecl},
		"endswith":   {Func: endswithFunc, Decl: endswithDecl},
		"trim":       {Func: trimFunc, Decl: trimDecl},
		"replace":    {Func: replaceFunc, Decl: replaceDecl},

		"str2num":  {Func: BuiltinFunc(str2numFunc), Decl: str2numDecl},
		"str2bool": {Func: BuiltinFunc(str2boolFunc), Decl: str2boolDecl},

		"len": {Func: BuiltinFunc(lenFunc), Decl: lenDecl},
		"has": {Func: BuiltinFunc(hasFunc), Decl: hasDecl},
		"del": {Func: BuiltinFunc(delFunc), Decl: delDecl},

		"sleep": {Func: sleepFunc(rt.Sleep), Decl: sleepDecl},

		"rand":  {Func: BuiltinFunc(randFunc), Decl: randDecl},
		"rand1": {Func: BuiltinFunc(rand1Func), Decl: rand1Decl},

		"move":   xyBuiltin("move", rt.Graphics.Move, rt.Print),
		"line":   xyBuiltin("line", rt.Graphics.Line, rt.Print),
		"rect":   xyBuiltin("rect", rt.Graphics.Rect, rt.Print),
		"circle": numBuiltin("circle", rt.Graphics.Circle, rt.Print),
		"width":  numBuiltin("width", rt.Graphics.Width, rt.Print),
		"color":  stringBuiltin("color", rt.Graphics.Color, rt.Print),
		"colour": stringBuiltin("colour", rt.Graphics.Color, rt.Print),
	}
	xyParams := []*parser.Var{
		{Name: "x", T: parser.NUM_TYPE},
		{Name: "y", T: parser.NUM_TYPE},
	}
	stringParam := []*parser.Var{{Name: "s", T: parser.STRING_TYPE}}
	numParam := []*parser.Var{{Name: "n", T: parser.NUM_TYPE}}
	inputParams := []*parser.Var{
		{Name: "id", T: parser.STRING_TYPE},
		{Name: "val", T: parser.STRING_TYPE},
	}
	eventHandlers := map[string]*parser.EventHandler{
		"down":    {Name: "down", Params: xyParams},
		"up":      {Name: "up", Params: xyParams},
		"move":    {Name: "move", Params: xyParams},
		"key":     {Name: "key", Params: stringParam},
		"input":   {Name: "input", Params: inputParams},
		"animate": {Name: "animate", Params: numParam},
	}
	globals := map[string]*parser.Var{
		"err":    {Name: "err", T: parser.BOOL_TYPE},
		"errmsg": {Name: "errmsg", T: parser.STRING_TYPE},
	}
	return Builtins{
		EventHandlers: eventHandlers,
		Funcs:         funcs,
		Print:         rt.Print,
		Globals:       globals,
	}
}

func DefaulParserBuiltins(rt *Runtime) parser.Builtins {
	builtins := DefaultBuiltins(rt)
	return newParserBuiltins(builtins)
}

type Runtime struct {
	Print    func(string)
	Read     func() string
	Sleep    func(dur time.Duration)
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

var readDecl = &parser.FuncDecl{
	Name:       "read",
	ReturnType: parser.STRING_TYPE,
}

func readFunc(readFn func() string) BuiltinFunc {
	return func(_ *scope, args []Value) Value {
		s := readFn()
		return &String{Val: s}
	}
}

var printDecl = &parser.FuncDecl{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) BuiltinFunc {
	return func(_ *scope, args []Value) Value {
		printFn(join(args, " ") + "\n")
		return nil
	}
}

func printfFunc(printFn func(string)) BuiltinFunc {
	return func(_ *scope, args []Value) Value {
		if len(args) < 1 {
			return newError("'printf' takes at least 1 argument")
		}
		format, ok := args[0].(*String)
		if !ok {
			return newError("first argument of 'printf' must be a string")
		}
		s := sprintf(format.Val, args[1:])
		printFn(s)
		return nil
	}
}

var sprintDecl = &parser.FuncDecl{
	Name:          "sprint",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.STRING_TYPE,
}

func sprintFunc(_ *scope, args []Value) Value {
	return &String{Val: join(args, " ")}
}

func sprintfFunc(_ *scope, args []Value) Value {
	if len(args) < 1 {
		return newError("'sprintf' takes at least 1 argument")
	}
	format, ok := args[0].(*String)
	if !ok {
		return newError("first argument of 'sprintf' must be a string")
	}
	return &String{Val: sprintf(format.Val, args[1:])}
}

func sprintf(s string, vals []Value) string {
	args := make([]any, len(vals))
	for i, val := range vals {
		args[i] = unwrapBasicValue(val)
	}
	return fmt.Sprintf(s, args...)
}

var joinDecl = &parser.FuncDecl{
	Name: "join",
	Params: []*parser.Var{
		{Name: "arr", T: parser.GENERIC_ARRAY},
		{Name: "sep", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func joinFunc(_ *scope, args []Value) Value {
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

func splitFunc(_ *scope, args []Value) Value {
	s := args[0].(*String)
	sep := args[1].(*String)
	slice := strings.Split(s.Val, sep.Val)
	elements := make([]Value, len(slice))
	for i, s := range slice {
		elements[i] = &String{Val: s}
	}
	return &Array{Elements: &elements}
}

var upperDecl = &parser.FuncDecl{
	Name: "upper",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func upperFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	return &String{Val: strings.ToUpper(s)}
}

var lowerDecl = &parser.FuncDecl{
	Name: "lower",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func lowerFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	return &String{Val: strings.ToLower(s)}
}

var indexDecl = &parser.FuncDecl{
	Name: "index",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "substr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NUM_TYPE,
}

func indexFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	substr := args[1].(*String).Val
	return &Num{Val: float64(strings.Index(s, substr))}
}

var startswithDecl = &parser.FuncDecl{
	Name: "startswith",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "startstr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func startswithFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	prefix := args[1].(*String).Val
	return &Bool{Val: strings.HasPrefix(s, prefix)}
}

var endswithDecl = &parser.FuncDecl{
	Name: "endswith",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "endstr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func endswithFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	suffix := args[1].(*String).Val
	return &Bool{Val: strings.HasSuffix(s, suffix)}
}

var trimDecl = &parser.FuncDecl{
	Name: "trim",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "cutset", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func trimFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	cutset := args[1].(*String).Val
	return &String{Val: strings.Trim(s, cutset)}
}

var replaceDecl = &parser.FuncDecl{
	Name: "replace",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "old", T: parser.STRING_TYPE},
		{Name: "new", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func replaceFunc(_ *scope, args []Value) Value {
	s := args[0].(*String).Val
	oldStr := args[1].(*String).Val
	newStr := args[2].(*String).Val
	return &String{Val: strings.ReplaceAll(s, oldStr, newStr)}
}

var str2numDecl = &parser.FuncDecl{
	Name:       "str2num",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func str2numFunc(scope *scope, args []Value) Value {
	resetGlobalErr(scope)
	s := args[0].(*String)
	n, err := strconv.ParseFloat(s.Val, 64)
	if err != nil {
		setGlobalErr(scope, "str2num: cannot parse "+s.Val)
	}
	return &Num{Val: n}
}

var str2boolDecl = &parser.FuncDecl{
	Name:       "str2bool",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.BOOL_TYPE,
}

func str2boolFunc(scope *scope, args []Value) Value {
	resetGlobalErr(scope)
	s := args[0].(*String)
	b, err := strconv.ParseBool(s.Val)
	if err != nil {
		setGlobalErr(scope, "str2bool: cannot parse "+s.Val)
	}
	return &Bool{Val: b}
}

func setGlobalErr(scope *scope, msg string) {
	globalErr(scope, true, msg)
}

func resetGlobalErr(scope *scope) {
	globalErr(scope, false, "")
}

func globalErr(scope *scope, isErr bool, msg string) {
	val, ok := scope.get("err")
	if !ok {
		panic("cannot find global err")
	}
	val.Set(&Bool{Val: isErr})
	val, ok = scope.get("errmsg")
	if !ok {
		panic("cannot find global errmsg")
	}
	val.Set(&String{Val: msg})
}

var lenDecl = &parser.FuncDecl{
	Name:       "len",
	Params:     []*parser.Var{{Name: "a", T: parser.ANY_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func lenFunc(_ *scope, args []Value) Value {
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

func hasFunc(_ *scope, args []Value) Value {
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

func delFunc(_ *scope, args []Value) Value {
	m := args[0].(*Map)
	keyStr := args[1].(*String)
	m.Delete(keyStr.Val)
	return nil
}

var sleepDecl = &parser.FuncDecl{
	Name:       "sleep",
	Params:     []*parser.Var{{Name: "seconds", T: parser.NUM_TYPE}},
	ReturnType: parser.NONE_TYPE,
}

func sleepFunc(sleepFn func(time.Duration)) BuiltinFunc {
	return func(_ *scope, args []Value) Value {
		secs := args[0].(*Num)
		dur := time.Duration(secs.Val * float64(time.Second))
		sleepFn(dur)
		return nil
	}
}

var randDecl = &parser.FuncDecl{
	Name:       "rand",
	Params:     []*parser.Var{{Name: "upper", T: parser.NUM_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func randFunc(_ *scope, args []Value) Value {
	upper := int32(args[0].(*Num).Val)
	return &Num{Val: float64(rand.Int31n(upper))} //nolint: gosec
}

var rand1Decl = &parser.FuncDecl{
	Name:       "rand",
	Params:     []*parser.Var{},
	ReturnType: parser.NUM_TYPE,
}

func rand1Func(_ *scope, args []Value) Value {
	return &Num{Val: rand.Float64()} //nolint: gosec
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
	result.Func = func(_ *scope, args []Value) Value {
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
	result.Func = func(_ *scope, args []Value) Value {
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
	result.Func = func(_ *scope, args []Value) Value {
		str := args[0].(*String)
		fn(str.Val)
		return nil
	}
	return result
}

func notImplementedFunc(name string, printFn func(string)) BuiltinFunc {
	return func(_ *scope, args []Value) Value {
		printFn("'" + name + "' not yet implemented\n")
		return nil
	}
}
