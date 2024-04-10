package evaluator

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"evylang.dev/evy/pkg/parser"
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

// BuiltinDecls returns the signatures of all built-in functions and
// event handlers, as well as predefined global variables, for use by
// the [parser.Parse] function.
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

type builtinFunc func(scope *scope, args []value) (value, error)

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
	return func(_ *scope, _ []value) (value, error) {
		s := readFn()
		return &stringVal{V: s}, nil
	}
}

func clsFunc(clsFn func()) builtinFunc {
	return func(_ *scope, _ []value) (value, error) {
		clsFn()
		return &noneVal{}, nil
	}
}

var printDecl = &parser.FuncDefStmt{
	Name:          "print",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func printFunc(printFn func(string)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		printFn(join(args, " ") + "\n")
		return &noneVal{}, nil
	}
}

func printfFunc(printFn func(string)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf(`%w: "printf" takes at least 1 argument`, ErrBadArguments)
		}
		format, ok := args[0].(*anyVal).V.(*stringVal)
		if !ok {
			return nil, fmt.Errorf(`%w: first argument of "printf" must be a string`, ErrBadArguments)
		}
		s := sprintf(format.V, args[1:])
		printFn(s)
		return &noneVal{}, nil
	}
}

var sprintDecl = &parser.FuncDefStmt{
	Name:          "sprint",
	VariadicParam: &parser.Var{Name: "a", T: parser.ANY_TYPE},
	ReturnType:    parser.STRING_TYPE,
}

func sprintFunc(_ *scope, args []value) (value, error) {
	return &stringVal{V: join(args, " ")}, nil
}

func sprintfFunc(_ *scope, args []value) (value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf(`%w: "sprintf" takes at least 1 argument`, ErrBadArguments)
	}
	format, ok := args[0].(*anyVal).V.(*stringVal)
	if !ok {
		return nil, fmt.Errorf(`%w: first argument of "sprintf" must be a string`, ErrBadArguments)
	}
	return &stringVal{V: sprintf(format.V, args[1:])}, nil
}

func sprintf(s string, vals []value) string {
	args := make([]any, len(vals))
	for i, val := range vals {
		args[i] = unwrapBasicvalue(val)
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

func joinFunc(_ *scope, args []value) (value, error) {
	arr := args[0].(*arrayVal)
	sep := args[1].(*stringVal)
	s := join(*arr.Elements, sep.V)
	return &stringVal{V: s}, nil
}

func join(args []value, sep string) string {
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

func splitFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal)
	sep := args[1].(*stringVal)
	slice := strings.Split(s.V, sep.V)
	elements := make([]value, len(slice))
	for i, s := range slice {
		elements[i] = &stringVal{V: s}
	}
	return &arrayVal{Elements: &elements}, nil
}

var upperDecl = &parser.FuncDefStmt{
	Name: "upper",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func upperFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	return &stringVal{V: strings.ToUpper(s)}, nil
}

var lowerDecl = &parser.FuncDefStmt{
	Name: "lower",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func lowerFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	return &stringVal{V: strings.ToLower(s)}, nil
}

var indexDecl = &parser.FuncDefStmt{
	Name: "index",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "substr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NUM_TYPE,
}

func indexFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	substr := args[1].(*stringVal).V
	return &numVal{V: float64(strings.Index(s, substr))}, nil
}

var startswithDecl = &parser.FuncDefStmt{
	Name: "startswith",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "startstr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func startswithFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	prefix := args[1].(*stringVal).V
	return &boolVal{V: strings.HasPrefix(s, prefix)}, nil
}

var endswithDecl = &parser.FuncDefStmt{
	Name: "endswith",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "endstr", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func endswithFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	suffix := args[1].(*stringVal).V
	return &boolVal{V: strings.HasSuffix(s, suffix)}, nil
}

var trimDecl = &parser.FuncDefStmt{
	Name: "trim",
	Params: []*parser.Var{
		{Name: "s", T: parser.STRING_TYPE},
		{Name: "cutset", T: parser.STRING_TYPE},
	},
	ReturnType: parser.STRING_TYPE,
}

func trimFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	cutset := args[1].(*stringVal).V
	return &stringVal{V: strings.Trim(s, cutset)}, nil
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

func replaceFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	oldStr := args[1].(*stringVal).V
	newStr := args[2].(*stringVal).V
	return &stringVal{V: strings.ReplaceAll(s, oldStr, newStr)}, nil
}

var str2numDecl = &parser.FuncDefStmt{
	Name:       "str2num",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func str2numFunc(scope *scope, args []value) (value, error) {
	resetGlobalErr(scope)
	s := args[0].(*stringVal)
	n, err := strconv.ParseFloat(s.V, 64)
	if err != nil {
		msg := fmt.Sprintf("str2num: cannot parse %q", s.V)
		setGlobalErr(scope, msg)
	}
	return &numVal{V: n}, nil
}

var str2boolDecl = &parser.FuncDefStmt{
	Name:       "str2bool",
	Params:     []*parser.Var{{Name: "s", T: parser.STRING_TYPE}},
	ReturnType: parser.BOOL_TYPE,
}

func str2boolFunc(scope *scope, args []value) (value, error) {
	resetGlobalErr(scope)
	s := args[0].(*stringVal)
	b, err := strconv.ParseBool(s.V)
	if err != nil {
		msg := fmt.Sprintf("str2bool: cannot parse %q", s.V)
		setGlobalErr(scope, msg)
	}
	return &boolVal{V: b}, nil
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
	val.Set(&boolVal{V: isErr})
	val, ok = scope.get("errmsg")
	if !ok {
		panic("cannot find global errmsg")
	}
	val.Set(&stringVal{V: msg})
}

var typeofDecl = &parser.FuncDefStmt{
	Name:       "typeof",
	Params:     []*parser.Var{{Name: "a", T: parser.ANY_TYPE}},
	ReturnType: parser.STRING_TYPE,
}

func typeofFunc(_ *scope, args []value) (value, error) {
	t := args[0].(*anyVal).T
	return &stringVal{V: t.String()}, nil
}

var lenDecl = &parser.FuncDefStmt{
	Name:       "len",
	Params:     []*parser.Var{{Name: "a", T: parser.ANY_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

func lenFunc(_ *scope, args []value) (value, error) {
	switch arg := args[0].(*anyVal).V.(type) {
	case *mapVal:
		return &numVal{V: float64(len(arg.Pairs))}, nil
	case *arrayVal:
		return &numVal{V: float64(len(*arg.Elements))}, nil
	case *stringVal:
		return &numVal{V: float64(len(arg.runes()))}, nil
	}
	return nil, fmt.Errorf(`%w: "len" takes 1 argument of type "string", array "[]" or map "{}" not %s`, ErrBadArguments, args[0].(*anyVal).T)
}

var hasDecl = &parser.FuncDefStmt{
	Name: "has",
	Params: []*parser.Var{
		{Name: "m", T: parser.GENERIC_MAP},
		{Name: "key", T: parser.STRING_TYPE},
	},
	ReturnType: parser.BOOL_TYPE,
}

func hasFunc(_ *scope, args []value) (value, error) {
	m := args[0].(*mapVal)
	key := args[1].(*stringVal)
	_, ok := m.Pairs[key.V]
	return &boolVal{V: ok}, nil
}

var delDecl = &parser.FuncDefStmt{
	Name: "del",
	Params: []*parser.Var{
		{Name: "m", T: parser.GENERIC_MAP},
		{Name: "key", T: parser.STRING_TYPE},
	},
	ReturnType: parser.NONE_TYPE,
}

func delFunc(_ *scope, args []value) (value, error) {
	m := args[0].(*mapVal)
	keyStr := args[1].(*stringVal)
	m.Delete(keyStr.V)
	return &noneVal{}, nil
}

var sleepDecl = &parser.FuncDefStmt{
	Name:       "sleep",
	Params:     []*parser.Var{{Name: "seconds", T: parser.NUM_TYPE}},
	ReturnType: parser.NONE_TYPE,
}

func sleepFunc(sleepFn func(time.Duration)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		secs := args[0].(*numVal)
		dur := time.Duration(secs.V * float64(time.Second))
		sleepFn(dur)
		return &noneVal{}, nil
	}
}

func exitFunc(_ *scope, args []value) (value, error) {
	return nil, ExitError(args[0].(*numVal).V)
}

func panicFunc(_ *scope, args []value) (value, error) {
	s := args[0].(*stringVal).V
	return nil, PanicError(s)
}

var randDecl = &parser.FuncDefStmt{
	Name:       "rand",
	Params:     []*parser.Var{{Name: "upper", T: parser.NUM_TYPE}},
	ReturnType: parser.NUM_TYPE,
}

// We need to manually seed for tinygo 0.28.1.
var randsource = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

func randFunc(_ *scope, args []value) (value, error) {
	upper := args[0].(*numVal).V
	if upper <= 0 || upper > 2147483647 { // [1, 2^31-1]
		return nil, fmt.Errorf(`%w: "rand %0.f" not in range 1 to 2147483647`, ErrBadArguments, upper)
	}
	return &numVal{V: float64(randsource.Int31n(int32(upper)))}, nil
}

var rand1Decl = &parser.FuncDefStmt{
	Name:       "rand",
	Params:     []*parser.Var{},
	ReturnType: parser.NUM_TYPE,
}

func rand1Func(_ *scope, _ []value) (value, error) {
	return &numVal{V: randsource.Float64()}, nil
}

var clearDecl = &parser.FuncDefStmt{
	Name:          "clear",
	VariadicParam: &parser.Var{Name: "s", T: parser.STRING_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func clearFunc(clearFn func(string)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf(`%w: "clear" takes 0 or 1 string arguments`, ErrBadArguments)
		}
		color := ""
		if len(args) == 1 {
			color = args[0].(*stringVal).V
		}
		clearFn(color)
		return &noneVal{}, nil
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
	return func(_ *scope, args []value) (value, error) {
		unit := args[0].(*numVal)
		color := args[1].(*stringVal)
		gridnFn(unit.V, color.V)
		return nil, nil
	}
}

func gridFunc(gridnFn func(float64, string)) builtinFunc {
	return func(_ *scope, _ []value) (value, error) {
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
	return func(_ *scope, args []value) (value, error) {
		vertices := make([][]float64, len(args))
		for i, arg := range args {
			vertex := arg.(*arrayVal)
			elements := *vertex.Elements
			if len(elements) != 2 {
				return nil, fmt.Errorf(`%w: "poly" argument %d has %d elements, expected 2 (x, y)`, ErrBadArguments, i+1, len(elements))
			}
			x := elements[0].(*numVal).V
			y := elements[1].(*numVal).V
			vertices[i] = []float64{x, y}
		}
		polyFn(vertices)
		return &noneVal{}, nil
	}
}

var ellipseDecl = &parser.FuncDefStmt{
	Name:          "ellipse",
	VariadicParam: &parser.Var{Name: "n", T: parser.NUM_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func ellipseFunc(ellipseFn func(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		argLen := len(args)
		if argLen < 3 || argLen == 6 || argLen > 7 {
			return nil, fmt.Errorf(`%w: "ellipse" requires 3, 4, 5 or 7 arguments, found %d`, ErrBadArguments, argLen)
		}
		x := args[0].(*numVal).V
		y := args[1].(*numVal).V
		radiusX := args[2].(*numVal).V
		radiusY := radiusX
		rotation := 0.0
		startAngle := 0.0
		endAngle := 360.0
		if argLen > 3 {
			radiusY = args[3].(*numVal).V
		}
		if argLen > 4 {
			rotation = args[4].(*numVal).V
		}
		if argLen > 6 {
			startAngle = args[5].(*numVal).V
			endAngle = args[6].(*numVal).V
		}
		ellipseFn(x, y, radiusX, radiusY, rotation, startAngle, endAngle)
		return &noneVal{}, nil
	}
}

var dashDecl = &parser.FuncDefStmt{
	Name:          "dash",
	VariadicParam: &parser.Var{Name: "segments", T: parser.NUM_TYPE},
	ReturnType:    parser.NONE_TYPE,
}

func dashFunc(dashFn func([]float64)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		segments := make([]float64, len(args))
		for i, arg := range args {
			segments[i] = arg.(*numVal).V
		}
		dashFn(segments)
		return &noneVal{}, nil
	}
}

var fontDecl = &parser.FuncDefStmt{
	Name:       "font",
	Params:     []*parser.Var{{Name: "properties", T: &parser.Type{Name: parser.MAP, Sub: parser.ANY_TYPE}}},
	ReturnType: parser.NONE_TYPE,
}

func parseFontProps(arg *mapVal) (map[string]any, error) {
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
		switch v := val.(*anyVal).V.(type) {
		case *stringVal:
			if propType != "string" {
				return nil, fmt.Errorf("%w: expected property %q of type %s, found string", ErrBadArguments, key, propType)
			}
			s := v.V
			if (key == "align" && s != "left" && s != "center" && s != "right") ||
				(key == "baseline" && s != "top" && s != "middle" && s != "bottom" && s != "alphabetic") {
				return nil, fmt.Errorf(`%w: expected property %q to be "top", "middle" or "bottom", found %q`, ErrBadArguments, key, s)
			}
			props[key] = v.V
		case *numVal:
			if propType != "num" {
				return nil, fmt.Errorf("%w: expected property %q of type %s, found num", ErrBadArguments, key, propType)
			}
			n := v.V
			if (key == "size" || key == "weight") && n <= 0 {
				return nil, fmt.Errorf(`%w: expected property %q to be greater than 0`, ErrBadArguments, key)
			}
			props[key] = v.V
		default:
			return nil, fmt.Errorf("%w: expected property %q of type %s, found %s", ErrBadArguments, key, propType, val.(*anyVal).T)
		}
	}
	return props, nil
}

func fontFunc(fontFn func(map[string]any)) builtinFunc {
	return func(_ *scope, args []value) (value, error) {
		arg := args[0].(*mapVal)
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
	result.Func = func(_ *scope, args []value) (value, error) {
		x := args[0].(*numVal)
		y := args[1].(*numVal)
		fn(x.V, y.V)
		return &noneVal{}, nil
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
	result.Func = func(_ *scope, args []value) (value, error) {
		x := args[0].(*numVal)
		y := args[1].(*numVal)
		result := fn(x.V, y.V)
		return &numVal{V: result}, nil
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
	result.Func = func(_ *scope, args []value) (value, error) {
		n := args[0].(*numVal)
		fn(n.V)
		return &noneVal{}, nil
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
	result.Func = func(_ *scope, args []value) (value, error) {
		n := args[0].(*numVal)
		result := fn(n.V)
		return &numVal{V: result}, nil
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
	result.Func = func(_ *scope, args []value) (value, error) {
		str := args[0].(*stringVal)
		fn(str.V)
		return &noneVal{}, nil
	}
	return result
}
