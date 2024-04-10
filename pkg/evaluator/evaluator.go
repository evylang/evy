package evaluator

import (
	"errors"
	"fmt"
	"math"

	"evylang.dev/evy/pkg/lexer"
	"evylang.dev/evy/pkg/parser"
)

// The Evaluator can return the following sentinel errors:
//   - ErrStopped is returned when the program has been stopped externally.
//   - ErrPanic and errors wrapping ErrPanic report runtime errors, such as an index out of bounds error.
//   - ErrInternal and errors wrapping ErrInternal report internal errors of the evaluator or AST. These errors should not occur.
var (
	ErrStopped = errors.New("stopped")

	ErrPanic         = errors.New("panic")
	ErrBounds        = fmt.Errorf("%w: index out of bounds", ErrPanic)
	ErrRangevalue    = fmt.Errorf("%w: bad range value", ErrPanic)
	ErrMapKey        = fmt.Errorf("%w: no value for map key", ErrPanic)
	ErrSlice         = fmt.Errorf("%w: bad slice", ErrPanic)
	ErrBadArguments  = fmt.Errorf("%w: bad arguments", ErrPanic)
	ErrAnyConversion = fmt.Errorf("%w: error converting any to type", ErrPanic)
	ErrVarNotSet     = fmt.Errorf("%w: variable has not been set yet", ErrPanic)

	ErrInternal         = errors.New("internal error")
	ErrUnknownNode      = fmt.Errorf("%w: unknown AST node", ErrInternal)
	ErrType             = fmt.Errorf("%w: type error", ErrInternal)
	ErrRangeType        = fmt.Errorf("%w: bad range type", ErrInternal)
	ErrOperation        = fmt.Errorf("%w: unknown operation", ErrInternal)
	ErrAssignmentTarget = fmt.Errorf("%w: bad assignment target", ErrInternal)
)

// ExitError is returned by [Evaluator.Eval] if Evy's [builtin exit]
// function is called.
//
// [builtin exit]: https://github.com/evylang/evy/blob/main/docs/builtins.md#exit
type ExitError int

// Error implements the error interface and returns message containing the exit status.
func (e ExitError) Error() string {
	return fmt.Sprintf("exit %d", int(e))
}

// PanicError is returned by [Evaluator.Eval] if Evy's [builtin panic]
// function is called or a runtime error occurs.
//
// [builtin panic]: https://github.com/evylang/evy/blob/main/docs/builtins.md#panic
type PanicError string

// Error implements the error interface and returns the panic message.
func (e PanicError) Error() string {
	return string(e)
}

// Unwrap returns the ErrPanic sentinel error so that it can be used in
//
//	errors.Is(err, evaluator.ErrPanic)
func (e *PanicError) Unwrap() error {
	return ErrPanic
}

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
func NewEvaluator(rt Runtime) *Evaluator {
	builtins := newBuiltins(rt)
	scope := newScope()
	for _, global := range builtins.Globals {
		t := global.Type()
		z := zero(t)
		scope.set(global.Name, z)
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

	yielder       Yielder // Yield to give JavaScript/browser events a chance to run.
	builtins      builtins
	eventHandlers map[string]*parser.EventHandlerStmt

	scope  *scope // Current top of scope stack
	global *scope // Global scope
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
	builtins := builtinsDeclsFromBuiltins(e.builtins)
	prog, err := parser.Parse(input, builtins)
	if err != nil {
		return err
	}
	return e.Eval(prog)
}

// Yielder is a runtime-implemented mechanism that causes the
// evaluation process to periodically give up control to the runtime.
// The Yield method of the Yielder interface is called at the
// beginning of each evaluation step. This allows the runtime to
// handle external tasks, such as processing events. For a sample
// implementation, see the sleepingYielder of the browser environment
// in the pkg/wasm directory.
type Yielder interface {
	Yield()
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

func (e *Evaluator) eval(node parser.Node) (value, error) {
	if e.Stopped {
		return nil, ErrStopped
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
		return &numVal{V: node.Value}, nil
	case *parser.StringLiteral:
		return &stringVal{V: node.Value}, nil
	case *parser.BoolLiteral:
		return &boolVal{V: node.Value}, nil
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
		return &breakVal{}, nil
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
		return &noneVal{}, nil
	}
	return nil, fmt.Errorf("%w: %v", ErrUnknownNode, node)
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
		arg, err := valueFromAny(param.Type(), args[i])
		if err != nil {
			return newErr(param, err)
		}
		e.scope.set(param.Name, arg)
	}
	_, err := e.eval(eh.Body)
	return err
}

func (e *Evaluator) yield() {
	if e.yielder != nil {
		e.yielder.Yield()
	}
}

func (e *Evaluator) evalProgram(program *parser.Program) (value, error) {
	e.eventHandlers = program.EventHandlers
	e.EventHandlerNames = make([]string, 0, len(e.eventHandlers))
	for name := range e.eventHandlers {
		e.EventHandlerNames = append(e.EventHandlerNames, name)
	}
	return e.evalStatments(program.Statements)
}

func (e *Evaluator) evalStatments(statements []parser.Node) (value, error) {
	var result value
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
	e.scope.set(decl.Var.Name, copyOrRef(val))
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

func (e *Evaluator) evalAny(a *parser.Any) (value, error) {
	val, err := e.eval(a.Value)
	if err != nil {
		return nil, err
	}
	if _, ok := val.(*anyVal); ok {
		panic("nested any value " + a.String())
	}
	return &anyVal{V: val, T: a.Value.Type()}, nil
}

func (e *Evaluator) evalArrayLiteral(arr *parser.ArrayLiteral) (value, error) {
	elements, err := e.evalExprList(arr.Elements)
	if err != nil {
		return nil, err
	}
	return &arrayVal{Elements: &elements}, nil
}

func (e *Evaluator) evalMapLiteral(m *parser.MapLiteral) (value, error) {
	pairs := map[string]value{}
	for key, node := range m.Pairs {
		val, err := e.eval(node)
		if err != nil {
			return nil, err
		}
		pairs[key] = copyOrRef(val)
	}
	order := make([]string, len(m.Order))
	copy(order, m.Order)
	return &mapVal{Pairs: pairs, Order: &order}, nil
}

func (e *Evaluator) evalFunccall(funcCall *parser.FuncCall) (value, error) {
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
		e.scope.set(param.Name, args[i])
	}
	if fd.VariadicParam != nil {
		varArg := &arrayVal{Elements: &args}
		e.scope.set(fd.VariadicParam.Name, varArg)
	}

	funcResult, err := e.eval(fd.Body)
	if err != nil {
		return nil, err
	}
	if returnValalue, ok := funcResult.(*returnVal); ok {
		return returnValalue.V, nil
	}
	return &noneVal{}, nil
}

func (e *Evaluator) evalReturn(ret *parser.ReturnStmt) (value, error) {
	if ret.Value == nil {
		return &returnVal{}, nil
	}
	val, err := e.eval(ret.Value)
	if err != nil {
		return nil, err
	}
	return &returnVal{V: val}, nil
}

func (e *Evaluator) evalIf(i *parser.IfStmt) (value, error) {
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
	return &noneVal{}, nil
}

func (e *Evaluator) evalWhile(w *parser.WhileStmt) (value, error) {
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

func (e *Evaluator) evalFor(f *parser.ForStmt) (value, error) {
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
	return &noneVal{}, nil
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
	case *arrayVal:
		aRange := &arrayRange{array: v, cur: 0}
		if f.LoopVar != nil {
			aRange.loopVar = zero(f.LoopVar.Type())
			e.scope.set(f.LoopVar.Name, aRange.loopVar)
		}
		return aRange, nil
	case *stringVal:
		sRange := &stringRange{str: v, cur: 0}
		if f.LoopVar != nil {
			sRange.loopVar = &stringVal{}
			e.scope.set(f.LoopVar.Name, sRange.loopVar)
		}
		return sRange, nil
	case *mapVal:
		order := make([]string, len(*v.Order))
		copy(order, *v.Order)
		mapRange := &mapRange{mapValal: v, cur: 0, order: order}
		if f.LoopVar != nil {
			mapRange.loopVar = &stringVal{}
			e.scope.set(f.LoopVar.Name, mapRange.loopVar)
		}
		return mapRange, nil
	}
	return nil, newErr(f.Range, ErrRangeType)
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
		return nil, newErr(r, fmt.Errorf("%w: step cannot be 0, infinite loop", ErrRangevalue))
	}

	sRange := &stepRange{
		cur:  start,
		stop: stop,
		step: step,
	}
	if loopVar != nil {
		loopVarVal := &numVal{}
		e.scope.set(loopVar.Name, loopVarVal)
		sRange.loopVar = loopVarVal
	}
	return sRange, nil
}

func (e *Evaluator) numValal(n parser.Node) (float64, error) {
	v, err := e.eval(n)
	if err != nil {
		return 0, err
	}
	numValal, ok := v.(*numVal)
	if !ok {
		return 0, newErr(n, fmt.Errorf("%w: expected number, found %v", ErrType, v))
	}
	return numValal.V, nil
}

func (e *Evaluator) numValalWithDefault(n parser.Node, defaultVal float64) (float64, error) {
	if n == nil {
		return defaultVal, nil
	}
	return e.numValal(n)
}

func (e *Evaluator) evalConditionalBlock(condBlock *parser.ConditionalBlock) (value, bool, error) {
	e.pushScope()
	defer e.popScope()
	cond, err := e.eval(condBlock.Condition)
	if err != nil {
		return nil, false, err
	}
	boolCond, ok := cond.(*boolVal)
	if !ok {
		err := fmt.Errorf("%w: conditional not a bool", ErrType)
		return nil, false, newErr(condBlock.Condition, err)
	}
	if boolCond.V {
		val, err := e.eval(condBlock.Block)
		return val, true, err
	}
	return nil, false, nil
}

func (e *Evaluator) evalBlockStatment(block *parser.BlockStatement) (value, error) {
	return e.evalStatments(block.Statements)
}

func (e *Evaluator) evalVar(v *parser.Var) (value, error) {
	if val, ok := e.scope.get(v.Name); ok {
		return val, nil
	}
	return nil, newErr(v, fmt.Errorf("%w: %s", ErrVarNotSet, v.Name))
}

func (e *Evaluator) evalExprList(terms []parser.Node) ([]value, error) {
	result := make([]value, len(terms))

	for i, t := range terms {
		evaluated, err := e.eval(t)
		if err != nil {
			return nil, err
		}
		result[i] = copyOrRef(evaluated)
	}

	return result, nil
}

func (e *Evaluator) evalUnaryExpr(expr *parser.UnaryExpression) (value, error) {
	right, err := e.eval(expr.Right)
	if err != nil {
		return nil, err
	}
	op := expr.Op
	switch right := right.(type) {
	case *numVal:
		if op == parser.OP_MINUS {
			return &numVal{V: -right.V}, nil
		}
	case *boolVal:
		if op == parser.OP_BANG {
			return &boolVal{V: !right.V}, nil
		}
	}
	return nil, newErr(expr, fmt.Errorf("%w (unary): %v", ErrOperation, expr))
}

func (e *Evaluator) evalBinaryExpr(expr *parser.BinaryExpression) (value, error) {
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
		return &boolVal{V: left.Equals(right)}, nil
	}
	if op == parser.OP_NOT_EQ {
		return &boolVal{V: !left.Equals(right)}, nil
	}
	var val value
	switch l := left.(type) {
	case *numVal:
		val, err = evalBinaryNumExpr(op, l, right.(*numVal))
	case *stringVal:
		val, err = evalBinaryStringExpr(op, l, right.(*stringVal))
	case *boolVal:
		val, err = evalBinaryBoolExpr(op, l, right.(*boolVal))
	case *arrayVal:
		val, err = evalBinaryArrayExpr(op, l, right.(*arrayVal))
	default:
		err = fmt.Errorf("%w (binary): %v", ErrOperation, expr)
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func canShortCircuit(op parser.Operator, left value) bool {
	l, ok := left.(*boolVal)
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

func evalBinaryNumExpr(op parser.Operator, left, right *numVal) (value, error) {
	switch op {
	case parser.OP_PLUS:
		return &numVal{V: left.V + right.V}, nil
	case parser.OP_MINUS:
		return &numVal{V: left.V - right.V}, nil
	case parser.OP_ASTERISK:
		return &numVal{V: left.V * right.V}, nil
	case parser.OP_PERCENT:
		return &numVal{V: math.Mod(left.V, right.V)}, nil
	case parser.OP_SLASH:
		return &numVal{V: left.V / right.V}, nil
	case parser.OP_GT:
		return &boolVal{V: left.V > right.V}, nil
	case parser.OP_LT:
		return &boolVal{V: left.V < right.V}, nil
	case parser.OP_GTEQ:
		return &boolVal{V: left.V >= right.V}, nil
	case parser.OP_LTEQ:
		return &boolVal{V: left.V <= right.V}, nil
	}
	return nil, fmt.Errorf("%w (num): %v", ErrOperation, op.String())
}

func evalBinaryStringExpr(op parser.Operator, left, right *stringVal) (value, error) {
	switch op {
	case parser.OP_PLUS:
		return &stringVal{V: left.V + right.V}, nil
	case parser.OP_GT:
		return &boolVal{left.V > right.V}, nil
	case parser.OP_LT:
		return &boolVal{left.V < right.V}, nil
	case parser.OP_GTEQ:
		return &boolVal{left.V >= right.V}, nil
	case parser.OP_LTEQ:
		return &boolVal{left.V <= right.V}, nil
	}
	return nil, fmt.Errorf("%w (string): %v", ErrOperation, op.String())
}

func evalBinaryBoolExpr(op parser.Operator, left, right *boolVal) (value, error) {
	switch op {
	case parser.OP_AND:
		return &boolVal{V: left.V && right.V}, nil
	case parser.OP_OR:
		return &boolVal{V: left.V || right.V}, nil
	}
	return nil, fmt.Errorf("%w (bool): %v", ErrOperation, op.String())
}

func evalBinaryArrayExpr(op parser.Operator, left, right *arrayVal) (value, error) {
	if op != parser.OP_PLUS {
		return nil, fmt.Errorf("%w (array): %v", ErrOperation, op.String())
	}
	result := left.Copy()
	rightElemnts := *right.Copy().Elements
	*result.Elements = append(*result.Elements, rightElemnts...)
	return result, nil
}

func (e *Evaluator) evalTarget(node parser.Node) (value, error) {
	switch n := node.(type) {
	case *parser.Var:
		return e.evalVar(n)
	case *parser.IndexExpression:
		return e.evalIndexExpr(n, true /* forAssign */)
	case *parser.DotExpression:
		return e.evalDotExpr(n, true /* forAssign */)
	}
	return nil, newErr(node, fmt.Errorf("%w: %v", ErrAssignmentTarget, node))
}

func (e *Evaluator) evalIndexExpr(expr *parser.IndexExpression, forAssign bool) (value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	index, err := e.eval(expr.Index)
	if err != nil {
		return nil, err
	}

	var val value
	switch l := left.(type) {
	case *arrayVal:
		val, err = l.Index(index)
	case *stringVal:
		val, err = l.Index(index)
	case *mapVal:
		strIndex, ok := index.(*stringVal)
		if !ok {
			return nil, newErr(expr.Left, fmt.Errorf("%w: expected string for map index, found %v", ErrType, index))
		}
		if forAssign {
			l.InsertKey(strIndex.V, expr.Type())
		}
		val, err = l.Get(strIndex.V)
	default:
		err = fmt.Errorf("%w: expected array, string or map with index, found %v", ErrType, expr.Left.Type())
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalDotExpr(expr *parser.DotExpression, forAssign bool) (value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	m, ok := left.(*mapVal)
	if !ok {
		return nil, newErr(expr, fmt.Errorf(`%w: expected map before ".", found %v`, ErrType, left))
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

func (e *Evaluator) evalSliceExpr(expr *parser.SliceExpression) (value, error) {
	left, err := e.eval(expr.Left)
	if err != nil {
		return nil, err
	}
	var start, end value
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
	var val value
	switch left := left.(type) {
	case *arrayVal:
		val, err = left.Slice(start, end)
	case *stringVal:
		val, err = left.Slice(start, end)
	default:
		err = fmt.Errorf(`%w: expected string or array before "[", found %v`, ErrType, left)
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalTypeAssertion(ta *parser.TypeAssertion) (value, error) {
	left, err := e.eval(ta.Left)
	if err != nil {
		return nil, err
	}
	// The parser should have already validated that the type of Left
	// is `any`, but we check anyway just in case. This may be removed
	// later.
	a, ok := left.(*anyVal)
	if !ok {
		return nil, newErr(ta, fmt.Errorf("%w: not an any", ErrAnyConversion))
	}
	if !a.T.Equals(ta.T) {
		return nil, newErr(ta, fmt.Errorf("%w: expected %v, found %v", ErrAnyConversion, ta.T, a.T))
	}
	return a.V, nil
}

func (e *Evaluator) pushScope() {
	e.scope = newInnerScope(e.scope)
}

func (e *Evaluator) pushFuncScope() func() {
	s := e.scope
	e.scope = newInnerScope(e.global)
	return func() { e.scope = s }
}

func (e *Evaluator) popScope() {
	e.scope = e.scope.outer
}
