package evaluator

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"foxygo.at/evy/pkg/parser"
)

type Builtin struct {
	Func BuiltinFunc
	Decl *parser.FuncDeclStmt
}

type Builtins struct {
	Funcs         map[string]Builtin
	EventHandlers map[string]*parser.EventHandlerStmt
	Globals       map[string]*parser.Var
	Runtime       Runtime
}

func (b Builtins) ParserBuiltins() parser.Builtins {
	funcs := make(map[string]*parser.FuncDeclStmt, len(b.Funcs))
	for name, builtin := range b.Funcs {
		funcs[name] = builtin.Decl
	}
	return parser.Builtins{
		Funcs:         funcs,
		EventHandlers: b.EventHandlers,
		Globals:       b.Globals,
	}
}

type BuiltinFunc func(scope *scope, args []Value) (Value, error)

func (b BuiltinFunc) Type() ValueType { return BUILTIN }
func (b BuiltinFunc) String() string  { return "builtin function" }

func DefaultBuiltins(rt Runtime) Builtins {
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

		"min":   xyRetBuiltin("min", math.Min),
		"max":   xyRetBuiltin("max", math.Max),
		"floor": numRetBuiltin("floor", math.Floor),
		"ceil":  numRetBuiltin("ceil", math.Ceil),
		"round": numRetBuiltin("round", math.Round),
		"pow":   xyRetBuiltin("pow", math.Pow),
		"log":   numRetBuiltin("log", math.Log),
		"sqrt":  numRetBuiltin("sqrt", math.Sqrt),
		"sin":   numRetBuiltin("sin", math.Sin),
		"cos":   numRetBuiltin("cos", math.Cos),
		"atan2": xyRetBuiltin("atan2", math.Atan2),

		"move":   xyBuiltin("move", rt.Move),
		"line":   xyBuiltin("line", rt.Line),
		"rect":   xyBuiltin("rect", rt.Rect),
		"circle": numBuiltin("circle", rt.Circle),
		"width":  numBuiltin("width", rt.Width),
		"color":  stringBuiltin("color", rt.Color),
		"colour": stringBuiltin("colour", rt.Color),
		"clear":  {Func: clearFunc(rt.Clear), Decl: clearDecl},

		"poly":    {Func: polyFunc(rt.Poly), Decl: polyDecl},
		"ellipse": {Func: ellipseFunc(rt.Ellipse), Decl: ellipseDecl},

		"stroke":  stringBuiltin("stroke", rt.Stroke),
		"fill":    stringBuiltin("fill", rt.Fill),
		"dash":    {Func: dashFunc(rt.Dash), Decl: dashDecl},
		"linecap": stringBuiltin("linecap", rt.Linecap),

		"text":       stringBuiltin("text", rt.Text),
		"textsize":   numBuiltin("textsize", rt.Textsize),
		"font":       stringBuiltin("font", rt.Font),
		"fontfamily": stringBuiltin("fontfamily", rt.Fontfamily),
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
	eventHandlers := map[string]*parser.EventHandlerStmt{
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
		Globals:       globals,
		Runtime:       rt,
	}
}

type Runtime interface {
	GraphicsRuntime
	Print(string)
	Read() string
	Sleep(dur time.Duration)
	Yielder() Yielder
}

type GraphicsRuntime interface {
	Move(x, y float64)
	Line(x, y float64)
	Rect(dx, dy float64)
	Circle(radius float64)
	Width(w float64)
	Color(s string)
	Clear(color string)

	// advanced graphics functions
	Poly(vertices [][]float64)
	Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64)
	Stroke(s string)
	Fill(s string)
	Dash(segments []float64)
	Linecap(s string)
	Text(s string)
	Textsize(size float64)
	Font(s string)
	Fontfamily(s string)
}

type UnimplementedRuntime struct {
	print func(string)
}

func (rt *UnimplementedRuntime) Print(s string) {
	if rt.print != nil {
		rt.print(s)
	} else {
		print(s)
	}
}

func (rt *UnimplementedRuntime) Unimplemented(s string) {
	rt.Print(fmt.Sprintf("%q not implemented\n", s))
}

func (rt *UnimplementedRuntime) Read() string              { rt.Unimplemented("read"); return "" }
func (rt *UnimplementedRuntime) Sleep(_ time.Duration)     { rt.Unimplemented("sleep") }
func (rt *UnimplementedRuntime) Yielder() Yielder          { rt.Unimplemented("yielder"); return nil }
func (rt *UnimplementedRuntime) Move(x, y float64)         { rt.Unimplemented("move") }
func (rt *UnimplementedRuntime) Line(x, y float64)         { rt.Unimplemented("line") }
func (rt *UnimplementedRuntime) Rect(x, y float64)         { rt.Unimplemented("rect") }
func (rt *UnimplementedRuntime) Circle(r float64)          { rt.Unimplemented("circle") }
func (rt *UnimplementedRuntime) Width(w float64)           { rt.Unimplemented("width") }
func (rt *UnimplementedRuntime) Color(s string)            { rt.Unimplemented("color") }
func (rt *UnimplementedRuntime) Clear(color string)        { rt.Unimplemented("clear") }
func (rt *UnimplementedRuntime) Poly(vertices [][]float64) { rt.Unimplemented("poly") }
func (rt *UnimplementedRuntime) Stroke(s string)           { rt.Unimplemented("stroke") }
func (rt *UnimplementedRuntime) Fill(s string)             { rt.Unimplemented("fill") }
func (rt *UnimplementedRuntime) Dash(segments []float64)   { rt.Unimplemented("dash") }
func (rt *UnimplementedRuntime) Linecap(s string)          { rt.Unimplemented("linecap") }
func (rt *UnimplementedRuntime) Text(s string)             { rt.Unimplemented("text") }
func (rt *UnimplementedRuntime) Textsize(size float64)     { rt.Unimplemented("textsize") }
func (rt *UnimplementedRuntime) Font(s string)             { rt.Unimplemented("font") }
func (rt *UnimplementedRuntime) Fontfamily(s string)       { rt.Unimplemented("fontfamily") }
func (rt *UnimplementedRuntime) Ellipse(x, y, rX, rY, rotation, startAngle, endAngle float64) {
	rt.Unimplemented("ellipse")
}

var readDecl = &parser.FuncDeclStmt{
	Name:       "read",
	ReturnType: parser.STRING_TYPE,
}

func readFunc(readFn func() string) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		s := readFn()
		return &String{Val: s}, nil
	}
}

var printDecl = &parser.FuncDeclStmt{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		printFn(join(args, " ") + "\n")
		return nil, nil
	}
}

func printfFunc(printFn func(string)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("%w: 'printf' takes at least 1 argument", ErrBadArguments)
		}
		format, ok := args[0].(*String)
		if !ok {
			return nil, fmt.Errorf("%w: first argument of 'printf' must be a string", ErrBadArguments)
		}
		s := sprintf(format.Val, args[1:])
		printFn(s)
		return nil, nil
	}
}

var sprintDecl = &parser.FuncDeclStmt{
	Name:          "sprint",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.STRING_TYPE,
}

func sprintFunc(_ *scope, args []Value) (Value, error) {
	return &String{Val: join(args, " ")}, nil
}

func sprintfFunc(_ *scope, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("%w: 'sprintf' takes at least 1 argument", ErrBadArguments)
	}
	format, ok := args[0].(*String)
	if !ok {
		return nil, fmt.Errorf("%w: first argument of 'sprintf' must be a string", ErrBadArguments)
	}
	return &String{Val: sprintf(format.Val, args[1:])}, nil
}

func sprintf(s string, vals []Value) string {
	args := make([]any, len(vals))
	for i, val := range vals {
		args[i] = unwrapBasicValue(val)
	}
	return fmt.Sprintf(s, args...)
}

var joinDecl = &parser.FuncDeclStmt{
	Name: "join",
	Params: []*parser.Var{
		{Name: "arr", T: parser.GENERIC_ARRAY},
		{Name: "sep", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func joinFunc(_ *scope, args []Value) (Value, error) {
	arr := args[0].(*Array)
	sep := args[1].(*String)
	s := join(*arr.Elements, sep.Val)
	return &String{Val: s}, nil
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

var splitDecl = &parser.FuncDeclStmt{
	Name: "split",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "sep", T: parser.STRING_TYPE},
	},
	ReturnType: stringArrayType,
}

func splitFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String)
	sep := args[1].(*String)
	slice := strings.Split(s.Val, sep.Val)
	elements := make([]Value, len(slice))
	for i, s := range slice {
		elements[i] = &String{Val: s}
	}
	return &Array{Elements: &elements}, nil
}

var upperDecl = &parser.FuncDeclStmt{
	Name: "upper",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func upperFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	return &String{Val: strings.ToUpper(s)}, nil
}

var lowerDecl = &parser.FuncDeclStmt{
	Name: "lower",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func lowerFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	return &String{Val: strings.ToLower(s)}, nil
}

var indexDecl = &parser.FuncDeclStmt{
	Name: "index",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "substr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NUM_TYPE,
}

func indexFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	substr := args[1].(*String).Val
	return &Num{Val: float64(strings.Index(s, substr))}, nil
}

var startswithDecl = &parser.FuncDeclStmt{
	Name: "startswith",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "startstr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func startswithFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	prefix := args[1].(*String).Val
	return &Bool{Val: strings.HasPrefix(s, prefix)}, nil
}

var endswithDecl = &parser.FuncDeclStmt{
	Name: "endswith",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "endstr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func endswithFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	suffix := args[1].(*String).Val
	return &Bool{Val: strings.HasSuffix(s, suffix)}, nil
}

var trimDecl = &parser.FuncDeclStmt{
	Name: "trim",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "cutset", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func trimFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	cutset := args[1].(*String).Val
	return &String{Val: strings.Trim(s, cutset)}, nil
}

var replaceDecl = &parser.FuncDeclStmt{
	Name: "replace",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "old", T: parser.STRING_TYPE},
		{Name: "new", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func replaceFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	oldStr := args[1].(*String).Val
	newStr := args[2].(*String).Val
	return &String{Val: strings.ReplaceAll(s, oldStr, newStr)}, nil
}

var str2numDecl = &parser.FuncDeclStmt{
	Name:       "str2num",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func str2numFunc(scope *scope, args []Value) (Value, error) {
	resetGlobalErr(scope)
	s := args[0].(*String)
	n, err := strconv.ParseFloat(s.Val, 64)
	if err != nil {
		setGlobalErr(scope, "str2num: cannot parse "+s.Val)
	}
	return &Num{Val: n}, nil
}

var str2boolDecl = &parser.FuncDeclStmt{
	Name:       "str2bool",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.BOOL_TYPE,
}

func str2boolFunc(scope *scope, args []Value) (Value, error) {
	resetGlobalErr(scope)
	s := args[0].(*String)
	b, err := strconv.ParseBool(s.Val)
	if err != nil {
		setGlobalErr(scope, "str2bool: cannot parse "+s.Val)
	}
	return &Bool{Val: b}, nil
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

var lenDecl = &parser.FuncDeclStmt{
	Name:       "len",
	Params:     []*parser.Var{{Name: "a", T: parser.ANY_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func lenFunc(_ *scope, args []Value) (Value, error) {
	switch arg := args[0].(type) {
	case *Map:
		return &Num{Val: float64(len(arg.Pairs))}, nil
	case *Array:
		return &Num{Val: float64(len(*arg.Elements))}, nil
	case *String:
		return &Num{Val: float64(len(arg.Val))}, nil
	}
	return nil, fmt.Errorf("%w: 'len' takes 1 argument of type 'string', array '[]' or map '{}' not %s", ErrBadArguments, args[0].Type())
}

var hasDecl = &parser.FuncDeclStmt{
	Name: "has",
	Params: []*parser.Var{
		{Name: "m", T: parser.GENERIC_MAP},
		{Name: "key", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func hasFunc(_ *scope, args []Value) (Value, error) {
	m := args[0].(*Map)
	key := args[1].(*String)
	_, ok := m.Pairs[key.Val]
	return &Bool{Val: ok}, nil
}

var delDecl = &parser.FuncDeclStmt{
	Name: "del",
	Params: []*parser.Var{
		{Name: "m", T: parser.GENERIC_MAP},
		{Name: "key", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NONE_TYPE,
}

func delFunc(_ *scope, args []Value) (Value, error) {
	m := args[0].(*Map)
	keyStr := args[1].(*String)
	m.Delete(keyStr.Val)
	return nil, nil
}

var sleepDecl = &parser.FuncDeclStmt{
	Name:       "sleep",
	Params:     []*parser.Var{{Name: "seconds", T: parser.NUM_TYPE}},
	ReturnType: parser.NONE_TYPE,
}

func sleepFunc(sleepFn func(time.Duration)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		secs := args[0].(*Num)
		dur := time.Duration(secs.Val * float64(time.Second))
		sleepFn(dur)
		return nil, nil
	}
}

var randDecl = &parser.FuncDeclStmt{
	Name:       "rand",
	Params:     []*parser.Var{{Name: "upper", T: parser.NUM_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func randFunc(_ *scope, args []Value) (Value, error) {
	upper := args[0].(*Num).Val
	if upper <= 0 || upper > 2147483647 { // [1, 2^31-1]
		return nil, fmt.Errorf(`%w: "rand %0.f" not in range 1 to 2147483647`, ErrBadArguments, upper)
	}
	return &Num{Val: float64(rand.Int31n(int32(upper)))}, nil //nolint: gosec
}

var rand1Decl = &parser.FuncDeclStmt{
	Name:       "rand",
	Params:     []*parser.Var{},
	ReturnType: parser.NUM_TYPE,
}

func rand1Func(_ *scope, args []Value) (Value, error) {
	return &Num{Val: rand.Float64()}, nil //nolint: gosec
}

var clearDecl = &parser.FuncDeclStmt{
	Name:          "clear",
	VariadicParam: &parser.Var{Name: "s", T: parser.STRING_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func clearFunc(clearFn func(string)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("%w: 'clear' takes 0 or 1 string arguments", ErrBadArguments)
		}
		color := ""
		if len(args) == 1 {
			color = args[0].(*String).Val
		}
		clearFn(color)
		return nil, nil
	}
}

var numArrayType *parser.Type = &parser.Type{
	Name: parser.ARRAY,
	Sub:  parser.NUM_TYPE,
}

var polyDecl = &parser.FuncDeclStmt{
	Name:          "poly",
	VariadicParam: &parser.Var{Name: "vertices", T: numArrayType},
	ReturnType:    parser.NONE_TYPE,
}

func polyFunc(polyFn func([][]float64)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		vertices := make([][]float64, len(args))
		for i, arg := range args {
			vertex := arg.(*Array)
			elements := *vertex.Elements
			if len(elements) != 2 {
				return nil, fmt.Errorf("%w: 'poly' argument %d has %d elements, expected 2 (x, y)", ErrBadArguments, i+1, len(elements))
			}
			x := elements[0].(*Num).Val
			y := elements[1].(*Num).Val
			vertices[i] = []float64{x, y}
		}
		polyFn(vertices)
		return nil, nil
	}
}

var ellipseDecl = &parser.FuncDeclStmt{
	Name:          "ellipse",
	VariadicParam: &parser.Var{Name: "n", T: parser.NUM_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func ellipseFunc(ellipseFn func(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		argLen := len(args)
		if argLen < 3 || argLen == 6 || argLen > 7 {
			return nil, fmt.Errorf("%w: 'ellipse' requires 3, 4, 5 or 7 arguments, found %d", ErrBadArguments, argLen)
		}
		x := args[0].(*Num).Val
		y := args[1].(*Num).Val
		radiusX := args[2].(*Num).Val
		radiusY := radiusX
		rotation := 0.0
		startAngle := 0.0
		endAngle := 360.0
		if argLen > 3 {
			radiusY = args[3].(*Num).Val
		}
		if argLen > 4 {
			rotation = args[4].(*Num).Val
		}
		if argLen > 6 {
			startAngle = args[5].(*Num).Val
			endAngle = args[6].(*Num).Val
		}
		ellipseFn(x, y, radiusX, radiusY, rotation, startAngle, endAngle)
		return nil, nil
	}
}

var dashDecl = &parser.FuncDeclStmt{
	Name:          "dash",
	VariadicParam: &parser.Var{Name: "segments", T: parser.NUM_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func dashFunc(dashFn func([]float64)) BuiltinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		segments := make([]float64, len(args))
		for i, arg := range args {
			segments[i] = arg.(*Num).Val
		}
		dashFn(segments)
		return nil, nil
	}
}

func xyDecl(name string) *parser.FuncDeclStmt {
	return &parser.FuncDeclStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "x", T: parser.NUM_TYPE},
			{Name: "y", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func xyBuiltin(name string, fn func(x, y float64)) Builtin {
	result := Builtin{Decl: xyDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		x := args[0].(*Num)
		y := args[1].(*Num)
		fn(x.Val, y.Val)
		return nil, nil
	}
	return result
}

func xyRetDecl(name string) *parser.FuncDeclStmt {
	return &parser.FuncDeclStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "x", T: parser.NUM_TYPE},
			{Name: "y", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NUM_TYPE,
	}
}

func xyRetBuiltin(name string, fn func(x, y float64) float64) Builtin {
	result := Builtin{Decl: xyRetDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		x := args[0].(*Num)
		y := args[1].(*Num)
		result := fn(x.Val, y.Val)
		return &Num{Val: result}, nil
	}
	return result
}

func numDecl(name string) *parser.FuncDeclStmt {
	return &parser.FuncDeclStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "n", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func numBuiltin(name string, fn func(n float64)) Builtin {
	result := Builtin{Decl: numDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		n := args[0].(*Num)
		fn(n.Val)
		return nil, nil
	}
	return result
}

func numRetDecl(name string) *parser.FuncDeclStmt {
	return &parser.FuncDeclStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "n", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NUM_TYPE,
	}
}

func numRetBuiltin(name string, fn func(n float64) float64) Builtin {
	result := Builtin{Decl: numRetDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		n := args[0].(*Num)
		result := fn(n.Val)
		return &Num{Val: result}, nil
	}
	return result
}

func stringDecl(name string) *parser.FuncDeclStmt {
	return &parser.FuncDeclStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "str", T: parser.STRING_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func stringBuiltin(name string, fn func(str string)) Builtin {
	result := Builtin{Decl: stringDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		str := args[0].(*String)
		fn(str.Val)
		return nil, nil
	}
	return result
}
