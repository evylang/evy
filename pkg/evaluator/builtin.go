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

type builtin struct {
	Func builtinFunc
	Decl *parser.FuncDefStmt
}

type builtins struct {
	Funcs         map[string]builtin
	EventHandlers map[string]*parser.EventHandlerStmt
	Globals       map[string]*parser.Var
	Runtime       Runtime
}

func BuiltinDecls() parser.Builtins {
	b := newBuiltins(&UnimplementedRuntime{})
	return builtinsDeclsFromBuiltins(b)
}

func builtinsDeclsFromBuiltins(b builtins) parser.Builtins {
	funcs := make(map[string]*parser.FuncDefStmt, len(b.Funcs))
	for name, builtin := range b.Funcs {
		funcs[name] = builtin.Decl
	}
	return parser.Builtins{
		Funcs:         funcs,
		EventHandlers: b.EventHandlers,
		Globals:       b.Globals,
	}
}

type builtinFunc func(scope *scope, args []Value) (Value, error)

func newBuiltins(rt Runtime) builtins {
	funcs := map[string]builtin{
		"read":   {Func: readFunc(rt.Read), Decl: readDecl},
		"cls":    {Func: clsFunc(rt.Cls), Decl: emptyDecl("cls")},
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

		"str2num":  {Func: builtinFunc(str2numFunc), Decl: str2numDecl},
		"str2bool": {Func: builtinFunc(str2boolFunc), Decl: str2boolDecl},

		"typeof": {Func: builtinFunc(typeofFunc), Decl: typeofDecl},

		"len": {Func: builtinFunc(lenFunc), Decl: lenDecl},
		"has": {Func: builtinFunc(hasFunc), Decl: hasDecl},
		"del": {Func: builtinFunc(delFunc), Decl: delDecl},

		"sleep": {Func: sleepFunc(rt.Sleep), Decl: sleepDecl},
		"exit":  {Func: builtinFunc(exitFunc), Decl: numDecl("exit")},
		"panic": {Func: builtinFunc(panicFunc), Decl: stringDecl("panic")},

		"rand":  {Func: builtinFunc(randFunc), Decl: randDecl},
		"rand1": {Func: builtinFunc(rand1Func), Decl: rand1Decl},

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

		"move":   xybuiltin("move", rt.Move),
		"line":   xybuiltin("line", rt.Line),
		"rect":   xybuiltin("rect", rt.Rect),
		"circle": numBuiltin("circle", rt.Circle),

		"width":  numBuiltin("width", rt.Width),
		"color":  stringBuiltin("color", rt.Color),
		"colour": stringBuiltin("colour", rt.Color),

		"clear": {Func: clearFunc(rt.Clear), Decl: clearDecl},
		"grid":  {Func: gridFunc(rt.Gridn), Decl: emptyDecl("grid")},
		"gridn": {Func: gridnFunc(rt.Gridn), Decl: gridnDecl},

		"poly":    {Func: polyFunc(rt.Poly), Decl: polyDecl},
		"ellipse": {Func: ellipseFunc(rt.Ellipse), Decl: ellipseDecl},

		"stroke":  stringBuiltin("stroke", rt.Stroke),
		"fill":    stringBuiltin("fill", rt.Fill),
		"dash":    {Func: dashFunc(rt.Dash), Decl: dashDecl},
		"linecap": stringBuiltin("linecap", rt.Linecap),

		"text": stringBuiltin("text", rt.Text),
		"font": {Func: fontFunc(rt.Font), Decl: fontDecl},
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
	return builtins{
		EventHandlers: eventHandlers,
		Funcs:         funcs,
		Globals:       globals,
		Runtime:       rt,
	}
}

var readDecl = &parser.FuncDefStmt{
	Name:       "read",
	ReturnType: parser.STRING_TYPE,
}

func readFunc(readFn func() string) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		s := readFn()
		return &String{Val: s}, nil
	}
}

func clsFunc(clsFn func()) builtinFunc {
	return func(_ *scope, _ []Value) (Value, error) {
		clsFn()
		return &None{}, nil
	}
}

var printDecl = &parser.FuncDefStmt{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		printFn(join(args, " ") + "\n")
		return &None{}, nil
	}
}

func printfFunc(printFn func(string)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf(`%w: "printf" takes at least 1 argument`, ErrBadArguments)
		}
		format, ok := args[0].(*String)
		if !ok {
			return nil, fmt.Errorf(`%w: first argument of "printf" must be a string`, ErrBadArguments)
		}
		s := sprintf(format.Val, args[1:])
		printFn(s)
		return &None{}, nil
	}
}

var sprintDecl = &parser.FuncDefStmt{
	Name:          "sprint",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.STRING_TYPE,
}

func sprintFunc(_ *scope, args []Value) (Value, error) {
	return &String{Val: join(args, " ")}, nil
}

func sprintfFunc(_ *scope, args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf(`%w: "sprintf" takes at least 1 argument`, ErrBadArguments)
	}
	format, ok := args[0].(*String)
	if !ok {
		return nil, fmt.Errorf(`%w: first argument of "sprintf" must be a string`, ErrBadArguments)
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

var joinDecl = &parser.FuncDefStmt{
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

var splitDecl = &parser.FuncDefStmt{
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
	return &Array{Elements: &elements, T: stringArrayType}, nil
}

var upperDecl = &parser.FuncDefStmt{
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

var lowerDecl = &parser.FuncDefStmt{
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

var indexDecl = &parser.FuncDefStmt{
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

var startswithDecl = &parser.FuncDefStmt{
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

var endswithDecl = &parser.FuncDefStmt{
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

var trimDecl = &parser.FuncDefStmt{
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

var replaceDecl = &parser.FuncDefStmt{
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

var str2numDecl = &parser.FuncDefStmt{
	Name:       "str2num",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func str2numFunc(scope *scope, args []Value) (Value, error) {
	resetGlobalErr(scope)
	s := args[0].(*String)
	n, err := strconv.ParseFloat(s.Val, 64)
	if err != nil {
		msg := fmt.Sprintf("str2num: cannot parse %q", s.Val)
		setGlobalErr(scope, msg)
	}
	return &Num{Val: n}, nil
}

var str2boolDecl = &parser.FuncDefStmt{
	Name:       "str2bool",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.BOOL_TYPE,
}

func str2boolFunc(scope *scope, args []Value) (Value, error) {
	resetGlobalErr(scope)
	s := args[0].(*String)
	b, err := strconv.ParseBool(s.Val)
	if err != nil {
		msg := fmt.Sprintf("str2bool: cannot parse %q", s.Val)
		setGlobalErr(scope, msg)
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

var typeofDecl = &parser.FuncDefStmt{
	Name:       "typeof",
	Params:     []*parser.Var{{Name: "a", T: parser.ANY_TYPE}},
	ReturnType: parser.STRING_TYPE,
}

func typeofFunc(_ *scope, args []Value) (Value, error) {
	t := args[0].Type()
	if a, ok := args[0].(*Any); ok {
		t = a.Val.Type()
	}
	return &String{Val: t.String()}, nil
}

var lenDecl = &parser.FuncDefStmt{
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
		return &Num{Val: float64(len(arg.runes()))}, nil
	}
	return nil, fmt.Errorf(`%w: "len" takes 1 argument of type "string", array "[]" or map "{}" not %s`, ErrBadArguments, args[0].Type())
}

var hasDecl = &parser.FuncDefStmt{
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

var delDecl = &parser.FuncDefStmt{
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
	return &None{}, nil
}

var sleepDecl = &parser.FuncDefStmt{
	Name:       "sleep",
	Params:     []*parser.Var{{Name: "seconds", T: parser.NUM_TYPE}},
	ReturnType: parser.NONE_TYPE,
}

func sleepFunc(sleepFn func(time.Duration)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		secs := args[0].(*Num)
		dur := time.Duration(secs.Val * float64(time.Second))
		sleepFn(dur)
		return &None{}, nil
	}
}

func exitFunc(_ *scope, args []Value) (Value, error) {
	return nil, ExitError(args[0].(*Num).Val)
}

func panicFunc(_ *scope, args []Value) (Value, error) {
	s := args[0].(*String).Val
	return nil, PanicError(s)
}

var randDecl = &parser.FuncDefStmt{
	Name:       "rand",
	Params:     []*parser.Var{{Name: "upper", T: parser.NUM_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

// We need to manually seed for tinygo 0.28.1.
var randsource = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

func randFunc(_ *scope, args []Value) (Value, error) {
	upper := args[0].(*Num).Val
	if upper <= 0 || upper > 2147483647 { // [1, 2^31-1]
		return nil, fmt.Errorf(`%w: "rand %0.f" not in range 1 to 2147483647`, ErrBadArguments, upper)
	}
	return &Num{Val: float64(randsource.Int31n(int32(upper)))}, nil
}

var rand1Decl = &parser.FuncDefStmt{
	Name:       "rand",
	Params:     []*parser.Var{},
	ReturnType: parser.NUM_TYPE,
}

func rand1Func(_ *scope, _ []Value) (Value, error) {
	return &Num{Val: randsource.Float64()}, nil
}

var clearDecl = &parser.FuncDefStmt{
	Name:          "clear",
	VariadicParam: &parser.Var{Name: "s", T: parser.STRING_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func clearFunc(clearFn func(string)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf(`%w: "clear" takes 0 or 1 string arguments`, ErrBadArguments)
		}
		color := ""
		if len(args) == 1 {
			color = args[0].(*String).Val
		}
		clearFn(color)
		return &None{}, nil
	}
}

var gridnDecl = &parser.FuncDefStmt{
	Name: "gridn",
	Params: []*parser.Var{
		{Name: "unit", T: parser.NUM_TYPE},
		{Name: "color", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NONE_TYPE,
}

func gridnFunc(gridnFn func(float64, string)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		unit := args[0].(*Num)
		color := args[1].(*String)
		gridnFn(unit.Val, color.Val)
		return nil, nil
	}
}

func gridFunc(gridnFn func(float64, string)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		gridnFn(10, "hsl(0deg 100% 0% / 50%)")
		return nil, nil
	}
}

var numArrayType *parser.Type = &parser.Type{
	Name: parser.ARRAY,
	Sub:  parser.NUM_TYPE,
}

var polyDecl = &parser.FuncDefStmt{
	Name:          "poly",
	VariadicParam: &parser.Var{Name: "vertices", T: numArrayType},
	ReturnType:    parser.NONE_TYPE,
}

func polyFunc(polyFn func([][]float64)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		vertices := make([][]float64, len(args))
		for i, arg := range args {
			vertex := arg.(*Array)
			elements := *vertex.Elements
			if len(elements) != 2 {
				return nil, fmt.Errorf(`%w: "poly" argument %d has %d elements, expected 2 (x, y)`, ErrBadArguments, i+1, len(elements))
			}
			x := elements[0].(*Num).Val
			y := elements[1].(*Num).Val
			vertices[i] = []float64{x, y}
		}
		polyFn(vertices)
		return &None{}, nil
	}
}

var ellipseDecl = &parser.FuncDefStmt{
	Name:          "ellipse",
	VariadicParam: &parser.Var{Name: "n", T: parser.NUM_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func ellipseFunc(ellipseFn func(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		argLen := len(args)
		if argLen < 3 || argLen == 6 || argLen > 7 {
			return nil, fmt.Errorf(`%w: "ellipse" requires 3, 4, 5 or 7 arguments, found %d`, ErrBadArguments, argLen)
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
		return &None{}, nil
	}
}

var dashDecl = &parser.FuncDefStmt{
	Name:          "dash",
	VariadicParam: &parser.Var{Name: "segments", T: parser.NUM_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func dashFunc(dashFn func([]float64)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		segments := make([]float64, len(args))
		for i, arg := range args {
			segments[i] = arg.(*Num).Val
		}
		dashFn(segments)
		return &None{}, nil
	}
}

var fontDecl = &parser.FuncDefStmt{
	Name:       "font",
	Params:     []*parser.Var{{Name: "properties", T: parser.GENERIC_MAP}},
	ReturnType: parser.NONE_TYPE,
}

func parseFontProps(arg *Map) (map[string]any, error) {
	props := map[string]any{}
	propTypes := map[string]string{
		"family":        "string",
		"size":          "num",
		"weight":        "num",
		"style":         "string",
		"baseline":      "string",
		"align":         "string",
		"letterspacing": "num",
	}
	for key, val := range arg.Pairs {
		propType, ok := propTypes[key]
		if !ok {
			return nil, fmt.Errorf("%w: unknown property %q", ErrBadArguments, key)
		}
		if a, ok := val.(*Any); ok {
			val = a.Val
		}
		switch v := val.(type) {
		case *String:
			if propType != "string" {
				return nil, fmt.Errorf("%w: expected property %q of type %s, found string", ErrBadArguments, key, propType)
			}
			s := v.Val
			if (key == "align" && s != "left" && s != "center" && s != "right") ||
				(key == "baseline" && s != "top" && s != "middle" && s != "bottom" && s != "alphabetic") {
				return nil, fmt.Errorf(`%w: expected property %q to be "top", "middle" or "bottom", found %q`, ErrBadArguments, key, s)
			}
			props[key] = v.Val
		case *Num:
			if propType != "num" {
				return nil, fmt.Errorf("%w: expected property %q of type %s, found num", ErrBadArguments, key, propType)
			}
			n := v.Val
			if (key == "size" || key == "weight") && n <= 0 {
				return nil, fmt.Errorf(`%w: expected property %q to be greater than 0`, ErrBadArguments, key)
			}
			props[key] = v.Val
		default:
			return nil, fmt.Errorf("%w: expected property %q of type %s, found %s", ErrBadArguments, key, propType, v.Type().String())
		}
	}
	return props, nil
}

func fontFunc(fontFn func(map[string]any)) builtinFunc {
	return func(_ *scope, args []Value) (Value, error) {
		arg := args[0].(*Map)
		properties, err := parseFontProps(arg)
		if err != nil {
			return nil, err
		}
		fontFn(properties)
		return nil, nil
	}
}

func emptyDecl(name string) *parser.FuncDefStmt {
	return &parser.FuncDefStmt{
		Name:       name,
		ReturnType: parser.NONE_TYPE,
	}
}

func xyDecl(name string) *parser.FuncDefStmt {
	return &parser.FuncDefStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "x", T: parser.NUM_TYPE},
			{Name: "y", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func xybuiltin(name string, fn func(x, y float64)) builtin {
	result := builtin{Decl: xyDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		x := args[0].(*Num)
		y := args[1].(*Num)
		fn(x.Val, y.Val)
		return &None{}, nil
	}
	return result
}

func xyRetDecl(name string) *parser.FuncDefStmt {
	return &parser.FuncDefStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "x", T: parser.NUM_TYPE},
			{Name: "y", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NUM_TYPE,
	}
}

func xyRetBuiltin(name string, fn func(x, y float64) float64) builtin {
	result := builtin{Decl: xyRetDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		x := args[0].(*Num)
		y := args[1].(*Num)
		result := fn(x.Val, y.Val)
		return &Num{Val: result}, nil
	}
	return result
}

func numDecl(name string) *parser.FuncDefStmt {
	return &parser.FuncDefStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "n", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func numBuiltin(name string, fn func(n float64)) builtin {
	result := builtin{Decl: numDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		n := args[0].(*Num)
		fn(n.Val)
		return &None{}, nil
	}
	return result
}

func numRetDecl(name string) *parser.FuncDefStmt {
	return &parser.FuncDefStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "n", T: parser.NUM_TYPE},
		},
		ReturnType: parser.NUM_TYPE,
	}
}

func numRetBuiltin(name string, fn func(n float64) float64) builtin {
	result := builtin{Decl: numRetDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		n := args[0].(*Num)
		result := fn(n.Val)
		return &Num{Val: result}, nil
	}
	return result
}

func stringDecl(name string) *parser.FuncDefStmt {
	return &parser.FuncDefStmt{
		Name: name,
		Params: []*parser.Var{
			{Name: "str", T: parser.STRING_TYPE},
		},
		ReturnType: parser.NONE_TYPE,
	}
}

func stringBuiltin(name string, fn func(str string)) builtin {
	result := builtin{Decl: stringDecl(name)}
	result.Func = func(_ *scope, args []Value) (Value, error) {
		str := args[0].(*String)
		fn(str.Val)
		return &None{}, nil
	}
	return result
}
