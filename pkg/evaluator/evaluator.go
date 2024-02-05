package evaluator

import (
	"fmt"
	"math"

	"evylang.dev/evy/pkg/abi"
	"evylang.dev/evy/pkg/lexer"
	"evylang.dev/evy/pkg/parser"
)

// Error is an Evy evaluator error associated with a [lexer.Token] that
// points to a location in the Evy source code that caused the error.
type Error struct {
	err   error
	Token *lexer.Token
}

// Error implements the error interface and returns the wrapped error
// message prefixed with the source location.
func (e *Error) Error() string {
	return e.Token.Location() + ": " + e.err.Error()
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.err
}

func newErr(node parser.Node, err error) *Error {
	return &Error{Token: node.Token(), err: err}
}

// NewEvaluator creates a new Evaluator for a given [Runtime]. Runtimes
// target different environments, such as the browser or the command
// line.
func NewEvaluator(rt abi.Runtime) *Evaluator {
	builtins := abi.NewBuiltins(rt)
	scope := abi.NewScope()
	for _, global := range builtins.Globals {
		t := global.Type()
		z := abi.Zero(t)
		scope.Set(global.Name, z)
	}
	return &Evaluator{
		builtins: builtins,
		scope:    scope,
		global:   scope,
		yielder:  builtins.Runtime.Yielder(),
	}
}

// Evaluator is a tree-walking interpreter that directly interprets the
// AST using the Run and Eval methods. The HandleEvent method can be
// used to allow the evaluator to handle events. The Evaluator does not
// preprocess or compile the AST to an intermediate representation,
// which results in a straightforward implementation that trades off
// execution performance for simplicity.
type Evaluator struct {
	// Stopped flags the evaluation to be stopped.
	// The unsynchronized access to the Stopped field is safe in WASM because
	// WASM is a single-threaded environment. TinyGo does not currently support
	// synchronization in reactor mode, see
	// https://github.com/tinygo-org/tinygo/issues/2735.
	Stopped           bool
	EventHandlerNames []string

	yielder       abi.Yielder // Yield to give JavaScript/browser events a chance to run.
	builtins      abi.Builtins
	eventHandlers map[string]*parser.EventHandlerStmt

	scope  *abi.Scope // Current top of scope stack
	global *abi.Scope // Global scope
}

// Event is a generic data structure that is passed to the
// [Evaluator] through the [Evaluator.HandleEvent] function. The
// evaluator tries to match the event's name and parameters to an event
// handler implementation in the Evy source code. If a matching handler
// is found, it is executed.
type Event struct {
	Name   string
	Params []any
}

// Run is a convenience function that parses and evaluates a given Evy
// source code input string. See the [Evaluator] type and
// [Evaluator.Eval] method for details on evaluation and errors.
func (e *Evaluator) Run(input string) error {
	builtins := abi.BuiltinsDeclsFromBuiltins(e.builtins)
	prog, err := parser.Parse(input, builtins)
	if err != nil {
		return err
	}
	return e.Eval(prog)
}

// Eval evaluates a [parser.Program], which is the root node of the AST.
// The program's statements are evaluated in order. If a runtime panic
// occurs, a wrapped [ErrPanic] is returned. If an internal error
// occurs, a wrapped [ErrInternal] is returned. Evaluation is also
// stopped if the built-in exit function is called, which results in an
// [ExitError]. If the evaluator's Stopped flag is externally set to
// true, evaluation is stopped and [ErrStopped] is returned.
func (e *Evaluator) Eval(prog *parser.Program) error {
	_, err := e.eval(prog)
	return err
}

func (e *Evaluator) eval(node parser.Node) (abi.Value, error) {
	if e.Stopped {
		return nil, abi.ErrStopped
	}
	e.yield()
	switch node := node.(type) {
	case *parser.Program:
		return e.evalProgram(node)
	case *parser.Decl:
		return nil, e.evalDecl(node)
	case *parser.TypedDeclStmt:
		return nil, e.evalDecl(node.Decl)
	case *parser.InferredDeclStmt:
		return nil, e.evalDecl(node.Decl)
	case *parser.AssignmentStmt:
		return nil, e.evalAssignment(node)
	case *parser.Var:
		return e.evalVar(node)
	case *parser.NumLiteral:
		return &abi.NumVal{V: node.Value}, nil
	case *parser.StringLiteral:
		return &abi.StringVal{V: node.Value}, nil
	case *parser.BoolLiteral:
		return &abi.BoolVal{V: node.Value}, nil
	case *parser.Any:
		return e.evalAny(node)
	case *parser.ArrayLiteral:
		return e.evalArrayLiteral(node)
	case *parser.MapLiteral:
		return e.evalMapLiteral(node)
	case *parser.FuncCall:
		return e.evalFunccall(node)
	case *parser.FuncCallStmt:
		return e.evalFunccall(node.FuncCall)
	case *parser.ReturnStmt:
		return e.evalReturn(node)
	case *parser.BreakStmt:
		return &abi.BreakVal{}, nil
	case *parser.IfStmt:
		return e.evalIf(node)
	case *parser.WhileStmt:
		return e.evalWhile(node)
	case *parser.ForStmt:
		return e.evalFor(node)
	case *parser.BlockStatement:
		return e.evalBlockStatment(node)
	case *parser.UnaryExpression:
		return e.evalUnaryExpr(node)
	case *parser.BinaryExpression:
		return e.evalBinaryExpr(node)
	case *parser.IndexExpression:
		return e.evalIndexExpr(node, false /* forAssign */)
	case *parser.SliceExpression:
		return e.evalSliceExpr(node)
	case *parser.DotExpression:
		return e.evalDotExpr(node, false /* forAssign */)
	case *parser.GroupExpression:
		return e.eval(node.Expr)
	case *parser.TypeAssertion:
		return e.evalTypeAssertion(node)
	case *parser.FuncDefStmt, *parser.EventHandlerStmt, *parser.EmptyStmt:
		return nil, nil
	}
	return nil, fmt.Errorf("%w: %v", abi.ErrUnknownNode, node)
}

// HandleEvent is called by an environment's event loop, passing the
// event ev to be handled. If the event's name and parameters match
// those of a predefined, built-in event handler signature, and if
// there is an event handler implementation in the Evy source code with
// the following signature:
//
//	on <event-name> [<event-params>]
//
// then the event handler implementation of the Evy source code is executed.
//
// For more details, see the [built-in documentation] on event handlers.
//
// [built-in documentation]: https://github.com/evylang/evy/blob/main/docs/builtins.md#event-handlers
func (e *Evaluator) HandleEvent(ev Event) error {
	eh := e.eventHandlers[ev.Name]
	if eh == nil {
		panic("no event handler for " + ev.Name)
	}
	restoreScope := e.pushFuncScope()
	defer restoreScope()
	args := ev.Params
	if len(args) < len(eh.Params) {
		panic("not enough arguments for " + ev.Name)
	}
	for i, param := range eh.Params {
		arg, err := abi.ValueFromAny(param.Type(), args[i])
		if err != nil {
			return newErr(param, err)
		}
		e.scope.Set(param.Name, arg)
	}
	_, err := e.eval(eh.Body)
	return err
}

func (e *Evaluator) yield() {
	if e.yielder != nil {
		e.yielder.Yield()
	}
}

func (e *Evaluator) evalProgram(program *parser.Program) (abi.Value, error) {
	e.eventHandlers = program.EventHandlers
	e.EventHandlerNames = make([]string, 0, len(e.eventHandlers))
	for name := range e.eventHandlers {
		e.EventHandlerNames = append(e.EventHandlerNames, name)
	}
	return e.evalStatments(program.Statements)
}

func (e *Evaluator) evalStatments(statements []parser.Node) (abi.Value, error) {
	var result abi.Value
	for _, statement := range statements {
		result, err := e.eval(statement)
		if err != nil {
			return nil, err
		}

		if isReturn(result) || isBreak(result) {
			return result, nil
		}
	}
	return result, nil
}

func (e *Evaluator) evalDecl(decl *parser.Decl) error {
	val, err := e.eval(decl.Value)
	if err != nil {
		return err
	}
	e.scope.Set(decl.Var.Name, abi.CopyOrRef(val))
	return nil
}

func (e *Evaluator) evalAssignment(assignment *parser.AssignmentStmt) error {
	val, err := e.eval(assignment.Value)
	if err != nil {
		return err
	}
	target, err := e.evalTarget(assignment.Target)
	if err != nil {
		return err
	}
	target.Set(val)
	return nil
}

func (e *Evaluator) evalAny(a *parser.Any) (abi.Value, error) {
	val, err := e.eval(a.Value)
	if err != nil {
		return nil, err
	}
	if _, ok := val.(*abi.AnyVal); ok {
		panic("nested any value " + a.String())
	}
	return &abi.AnyVal{V: val}, nil
}

func (e *Evaluator) evalArrayLiteral(arr *parser.ArrayLiteral) (abi.Value, error) {
	elements, err := e.evalExprList(arr.Elements)
	if err != nil {
		return nil, err
	}
	return &abi.ArrayVal{Elements: &elements, T: arr.T}, nil
}

func (e *Evaluator) evalMapLiteral(m *parser.MapLiteral) (abi.Value, error) {
	pairs := map[string]abi.Value{}
	for key, node := range m.Pairs {
		val, err := e.eval(node)
		if err != nil {
			return nil, err
		}
		pairs[key] = abi.CopyOrRef(val)
	}
	order := make([]string, len(m.Order))
	copy(order, m.Order)
	return &abi.MapVal{Pairs: pairs, Order: &order, T: m.T}, nil
}

func (e *Evaluator) evalFunccall(funcCall *parser.FuncCall) (abi.Value, error) {
	args, err := e.evalExprList(funcCall.Arguments)
	if err != nil {
		return nil, err
	}
	builtin, ok := e.builtins.Funcs[funcCall.Name]
	if ok {
		val, err := builtin.Func(e.scope, args)
		if err != nil {
			return nil, newErr(funcCall, err)
		}
		return val, nil
	}
	restoreScope := e.pushFuncScope()
	defer restoreScope()

	// Add func args to scope
	fd := funcCall.FuncDef
	for i, param := range fd.Params {
		e.scope.Set(param.Name, args[i])
	}
	if fd.VariadicParam != nil {
		varArg := &abi.ArrayVal{Elements: &args, T: fd.VariadicParamType}
		e.scope.Set(fd.VariadicParam.Name, varArg)
	}

	funcResult, err := e.eval(fd.Body)
	if err != nil {
		return nil, err
	}
	if returnValalue, ok := funcResult.(*abi.ReturnVal); ok {
		return returnValalue.V, nil
	}
	return nil, nil
}

func (e *Evaluator) evalReturn(ret *parser.ReturnStmt) (abi.Value, error) {
	if ret.Value == nil {
		return &abi.ReturnVal{}, nil
	}
	val, err := e.eval(ret.Value)
	if err != nil {
		return nil, err
	}
	return &abi.ReturnVal{V: val}, nil
}

func (e *Evaluator) evalIf(i *parser.IfStmt) (abi.Value, error) {
	val, ok, err := e.evalConditionalBlock(i.IfBlock)
	if err != nil {
		return nil, err
	}
	if ok {
		return val, nil
	}
	for _, elseif := range i.ElseIfBlocks {
		val, ok, err := e.evalConditionalBlock(elseif)
		if err != nil {
			return nil, err
		}
		if ok {
			return val, nil
		}
	}
	if i.Else != nil {
		e.pushScope()
		defer e.popScope()
		return e.eval(i.Else)
	}
	return nil, nil
}

func (e *Evaluator) evalWhile(w *parser.WhileStmt) (abi.Value, error) {
	whileBlock := &w.ConditionalBlock
	val, ok, err := e.evalConditionalBlock(whileBlock)
	for ok && err == nil && !isReturn(val) && !isBreak(val) {
		val, ok, err = e.evalConditionalBlock(whileBlock)
	}
	if isBreak(val) {
		val = nil
	}
	return val, err
}

func (e *Evaluator) evalFor(f *parser.ForStmt) (abi.Value, error) {
	e.pushScope()
	defer e.popScope()
	r, err := e.newRange(f)
	if err != nil {
		return nil, err
	}
	for r.next() {
		val, err := e.eval(f.Block)
		if err != nil {
			return nil, err
		}
		if isBreak(val) {
			return nil, nil
		}
		if isReturn(val) {
			return val, nil
		}
	}
	return nil, nil
}

func (e *Evaluator) newRange(f *parser.ForStmt) (ranger, error) {
	if r, ok := f.Range.(*parser.StepRange); ok {
		return e.newStepRange(r, f.LoopVar)
	}
	rangeVal, err := e.eval(f.Range)
	if err != nil {
		return nil, err
	}

	switch v := rangeVal.(type) {
	case *abi.ArrayVal:
		aRange := &arrayRange{array: v, cur: 0}
		if f.LoopVar != nil {
			aRange.loopVar = abi.Zero(f.LoopVar.Type())
			e.scope.Set(f.LoopVar.Name, aRange.loopVar)
		}
		return aRange, nil
	case *abi.StringVal:
		sRange := &stringRange{str: v, cur: 0}
		if f.LoopVar != nil {
			sRange.loopVar = &abi.StringVal{}
			e.scope.Set(f.LoopVar.Name, sRange.loopVar)
		}
		return sRange, nil
	case *abi.MapVal:
		order := make([]string, len(*v.Order))
		copy(order, *v.Order)
		mapRange := &mapRange{mapValal: v, cur: 0, order: order}
		if f.LoopVar != nil {
			mapRange.loopVar = &abi.StringVal{}
			e.scope.Set(f.LoopVar.Name, mapRange.loopVar)
		}
		return mapRange, nil
	}
	return nil, newErr(f.Range, abi.ErrRangeType)
}

func (e *Evaluator) newStepRange(r *parser.StepRange, loopVar *parser.Var) (ranger, error) {
	start, err := e.numValalWithDefault(r.Start, 0.0)
	if err != nil {
		return nil, err
	}
	stop, err := e.numValal(r.Stop)
	if err != nil {
		return nil, err
	}
	step, err := e.numValalWithDefault(r.Step, 1.0)
	if err != nil {
		return nil, err
	}
	if step == 0 {
		return nil, newErr(r, fmt.Errorf("%w: step cannot be 0, infinite loop", abi.ErrRangevalue))
	}

	sRange := &stepRange{
		cur:  start,
		stop: stop,
		step: step,
	}
	if loopVar != nil {
		loopVarVal := &abi.NumVal{}
		e.scope.Set(loopVar.Name, loopVarVal)
		sRange.loopVar = loopVarVal
	}
	return sRange, nil
}

func (e *Evaluator) numValal(n parser.Node) (float64, error) {
	v, err := e.eval(n)
	if err != nil {
		return 0, err
	}
	numValal, ok := v.(*abi.NumVal)
	if !ok {
		return 0, newErr(n, fmt.Errorf("%w: expected number, found %v", abi.ErrType, v))
	}
	return numValal.V, nil
}

func (e *Evaluator) numValalWithDefault(n parser.Node, defaultVal float64) (float64, error) {
	if n == nil {
		return defaultVal, nil
	}
	return e.numValal(n)
}

func (e *Evaluator) evalConditionalBlock(condBlock *parser.ConditionalBlock) (abi.Value, bool, error) {
	e.pushScope()
	defer e.popScope()
	cond, err := e.eval(condBlock.Condition)
	if err != nil {
		return nil, false, err
	}
	boolCond, ok := cond.(*abi.BoolVal)
	if !ok {
		err := fmt.Errorf("%w: conditional not a bool", abi.ErrType)
		return nil, false, newErr(condBlock.Condition, err)
	}
	if boolCond.V {
		val, err := e.eval(condBlock.Block)
		return val, true, err
	}
	return nil, false, nil
}

func (e *Evaluator) evalBlockStatment(block *parser.BlockStatement) (abi.Value, error) {
	return e.evalStatments(block.Statements)
}

func (e *Evaluator) evalVar(v *parser.Var) (abi.Value, error) {
	if val, ok := e.scope.Get(v.Name); ok {
		return val, nil
	}
	return nil, newErr(v, fmt.Errorf("%w: %s", abi.ErrVarNotSet, v.Name))
}

func (e *Evaluator) evalExprList(terms []parser.Node) ([]abi.Value, error) {
	result := make([]abi.Value, len(terms))

	for i, t := range terms {
		evaluated, err := e.eval(t)
		if err != nil {
			return nil, err
		}
		result[i] = abi.CopyOrRef(evaluated)
	}

	return result, nil
}

func (e *Evaluator) evalUnaryExpr(expr *parser.UnaryExpression) (abi.Value, error) {
	right, err := e.eval(expr.Right)
	if err != nil {
		return nil, err
	}
	op := expr.Op
	switch right := right.(type) {
	case *abi.NumVal:
		if op == parser.OP_MINUS {
			return &abi.NumVal{V: -right.V}, nil
		}
	case *abi.BoolVal:
		if op == parser.OP_BANG {
			return &abi.BoolVal{V: !right.V}, nil
		}
	}
	return nil, newErr(expr, fmt.Errorf("%w (unary): %v", abi.ErrOperation, expr))
}

func (e *Evaluator) evalBinaryExpr(expr *parser.BinaryExpression) (abi.Value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	// We need to short-circuit the evaluation of expr.Right for and/or
	// operators. We start of treating "right" as "left" and only if
	// we cannot short-circuit do we evaluate expr.Right. If we do
	// short-circuit, it does not matter what "right" is.
	right := left
	if !canShortCircuit(expr.Op, left) {
		right, err = e.eval(expr.Right)
		if err != nil {
			return nil, err
		}
	}
	op := expr.Op
	if op == parser.OP_EQ {
		return &abi.BoolVal{V: left.Equals(right)}, nil
	}
	if op == parser.OP_NOT_EQ {
		return &abi.BoolVal{V: !left.Equals(right)}, nil
	}
	var val abi.Value
	switch l := left.(type) {
	case *abi.NumVal:
		val, err = evalBinaryNumExpr(op, l, right.(*abi.NumVal))
	case *abi.StringVal:
		val, err = evalBinaryStringExpr(op, l, right.(*abi.StringVal))
	case *abi.BoolVal:
		val, err = evalBinaryBoolExpr(op, l, right.(*abi.BoolVal))
	case *abi.ArrayVal:
		val, err = evalBinaryArrayExpr(op, l, right.(*abi.ArrayVal), expr.Type())
	default:
		err = fmt.Errorf("%w (binary): %v", abi.ErrOperation, expr)
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func canShortCircuit(op parser.Operator, left abi.Value) bool {
	l, ok := left.(*abi.BoolVal)
	if !ok {
		return false
	}
	switch op {
	case parser.OP_AND:
		return !l.V // short-circuit AND when left is false
	case parser.OP_OR:
		return l.V // short-circuit OR when left is true
	}
	return false
}

func evalBinaryNumExpr(op parser.Operator, left, right *abi.NumVal) (abi.Value, error) {
	switch op {
	case parser.OP_PLUS:
		return &abi.NumVal{V: left.V + right.V}, nil
	case parser.OP_MINUS:
		return &abi.NumVal{V: left.V - right.V}, nil
	case parser.OP_ASTERISK:
		return &abi.NumVal{V: left.V * right.V}, nil
	case parser.OP_PERCENT:
		return &abi.NumVal{V: math.Mod(left.V, right.V)}, nil
	case parser.OP_SLASH:
		return &abi.NumVal{V: left.V / right.V}, nil
	case parser.OP_GT:
		return &abi.BoolVal{V: left.V > right.V}, nil
	case parser.OP_LT:
		return &abi.BoolVal{V: left.V < right.V}, nil
	case parser.OP_GTEQ:
		return &abi.BoolVal{V: left.V >= right.V}, nil
	case parser.OP_LTEQ:
		return &abi.BoolVal{V: left.V <= right.V}, nil
	}
	return nil, fmt.Errorf("%w (num): %v", abi.ErrOperation, op.String())
}

func evalBinaryStringExpr(op parser.Operator, left, right *abi.StringVal) (abi.Value, error) {
	switch op {
	case parser.OP_PLUS:
		return &abi.StringVal{V: left.V + right.V}, nil
	case parser.OP_GT:
		return &abi.BoolVal{V: left.V > right.V}, nil
	case parser.OP_LT:
		return &abi.BoolVal{V: left.V < right.V}, nil
	case parser.OP_GTEQ:
		return &abi.BoolVal{V: left.V >= right.V}, nil
	case parser.OP_LTEQ:
		return &abi.BoolVal{V: left.V <= right.V}, nil
	}
	return nil, fmt.Errorf("%w (string): %v", abi.ErrOperation, op.String())
}

func evalBinaryBoolExpr(op parser.Operator, left, right *abi.BoolVal) (abi.Value, error) {
	switch op {
	case parser.OP_AND:
		return &abi.BoolVal{V: left.V && right.V}, nil
	case parser.OP_OR:
		return &abi.BoolVal{V: left.V || right.V}, nil
	}
	return nil, fmt.Errorf("%w (bool): %v", abi.ErrOperation, op.String())
}

func evalBinaryArrayExpr(op parser.Operator, left, right *abi.ArrayVal, t *parser.Type) (abi.Value, error) {
	if op != parser.OP_PLUS {
		return nil, fmt.Errorf("%w (array): %v", abi.ErrOperation, op.String())
	}
	result := left.Copy()
	result.T = t
	rightElemnts := *right.Copy().Elements
	*result.Elements = append(*result.Elements, rightElemnts...)
	return result, nil
}

func (e *Evaluator) evalTarget(node parser.Node) (abi.Value, error) {
	switch n := node.(type) {
	case *parser.Var:
		return e.evalVar(n)
	case *parser.IndexExpression:
		return e.evalIndexExpr(n, true /* forAssign */)
	case *parser.DotExpression:
		return e.evalDotExpr(n, true /* forAssign */)
	}
	return nil, newErr(node, fmt.Errorf("%w: %v", abi.ErrAssignmentTarget, node))
}

func (e *Evaluator) evalIndexExpr(expr *parser.IndexExpression, forAssign bool) (abi.Value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	index, err := e.eval(expr.Index)
	if err != nil {
		return nil, err
	}

	var val abi.Value
	switch l := left.(type) {
	case *abi.ArrayVal:
		val, err = l.Index(index)
	case *abi.StringVal:
		val, err = l.Index(index)
	case *abi.MapVal:
		strIndex, ok := index.(*abi.StringVal)
		if !ok {
			return nil, newErr(expr.Left, fmt.Errorf("%w: expected string for map index, found %v", abi.ErrType, index))
		}
		if forAssign {
			l.InsertKey(strIndex.V, expr.Type())
		}
		val, err = l.Get(strIndex.V)
	default:
		err = fmt.Errorf("%w: expected array, string or map with index, found %v", abi.ErrType, left.Type())
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalDotExpr(expr *parser.DotExpression, forAssign bool) (abi.Value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	m, ok := left.(*abi.MapVal)
	if !ok {
		return nil, newErr(expr, fmt.Errorf(`%w: expected map before ".", found %v`, abi.ErrType, left))
	}
	if forAssign {
		m.InsertKey(expr.Key, expr.Type())
	}
	val, err := m.Get(expr.Key)
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalSliceExpr(expr *parser.SliceExpression) (abi.Value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	var start, end abi.Value
	if expr.Start != nil {
		if start, err = e.eval(expr.Start); err != nil {
			return nil, err
		}
	}
	if expr.End != nil {
		if end, err = e.eval(expr.End); err != nil {
			return nil, err
		}
	}
	var val abi.Value
	switch left := left.(type) {
	case *abi.ArrayVal:
		val, err = left.Slice(start, end)
	case *abi.StringVal:
		val, err = left.Slice(start, end)
	default:
		err = fmt.Errorf(`%w: expected string or array before "[", found %v`, abi.ErrType, left)
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalTypeAssertion(ta *parser.TypeAssertion) (abi.Value, error) {
	left, err := e.eval(ta.Left)
	if err != nil {
		return nil, err
	}
	// The parser should have already validated that the type of Left
	// is `any`, but we check anyway just in case. This may be removed
	// later.
	a, ok := left.(*abi.AnyVal)
	if !ok {
		return nil, newErr(ta, fmt.Errorf("%w: not an any", abi.ErrAnyConversion))
	}
	if !a.V.Type().Equals(ta.T) {
		return nil, newErr(ta, fmt.Errorf("%w: expected %v, found %v", abi.ErrAnyConversion, ta.T, a.V.Type()))
	}
	return a.V, nil
}

func (e *Evaluator) pushScope() {
	e.scope = abi.NewInnerScope(e.scope)
}

func (e *Evaluator) pushFuncScope() func() {
	s := e.scope
	e.scope = abi.NewInnerScope(e.global)
	return func() { e.scope = s }
}

func (e *Evaluator) popScope() {
	e.scope = e.scope.Outer
}

func isReturn(val abi.Value) bool {
	_, ok := val.(*abi.ReturnVal)
	return ok
}

func isBreak(val abi.Value) bool {
	_, ok := val.(*abi.BreakVal)
	return ok
}
