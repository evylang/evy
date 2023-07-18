// Package evaluator evaluates a given syntax tree as created by the
// parser packages. It also exports a Run and RunWithBuiltings function
// which creates and calls a Parser.
package evaluator

import (
	"errors"
	"fmt"
	"math"

	"foxygo.at/evy/pkg/lexer"
	"foxygo.at/evy/pkg/parser"
)

var (
	ErrStopped = errors.New("stopped")

	ErrRuntime       = errors.New("runtime error")
	ErrBounds        = fmt.Errorf("%w: index out of bounds", ErrRuntime)
	ErrRangeValue    = fmt.Errorf("%w: bad range value", ErrRuntime)
	ErrMapKey        = fmt.Errorf("%w: no value for map key", ErrRuntime)
	ErrSlice         = fmt.Errorf("%w: bad slice", ErrRuntime)
	ErrBadArguments  = fmt.Errorf("%w: bad arguments", ErrRuntime)
	ErrAnyConversion = fmt.Errorf("%w: error converting any to type", ErrRuntime)

	ErrInternal         = errors.New("internal error")
	ErrUnknownNode      = fmt.Errorf("%w: unknown AST node", ErrInternal)
	ErrType             = fmt.Errorf("%w: type error", ErrInternal)
	ErrRangeType        = fmt.Errorf("%w: bad range type", ErrInternal)
	ErrNoVarible        = fmt.Errorf("%w: no variable", ErrInternal)
	ErrOperation        = fmt.Errorf("%w: unknown operation", ErrInternal)
	ErrAssignmentTarget = fmt.Errorf("%w: bad assignment target", ErrInternal)
)

type ExitError int

func (e ExitError) Error() string {
	return fmt.Sprintf("exit %d", int(e))
}

// Error is an Evy evaluator error.
type Error struct {
	err   error
	token *lexer.Token
}

func (e *Error) Error() string {
	return e.token.Location() + ": " + e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func newErr(node parser.Node, err error) *Error {
	return &Error{token: node.Token(), err: err}
}

func NewEvaluator(builtins Builtins) *Evaluator {
	scope := newScope()
	for _, global := range builtins.Globals {
		t := global.Type()
		z := zero(t)
		scope.set(global.Name, z, t)
	}
	return &Evaluator{
		builtins: builtins,
		scope:    scope,
		global:   scope,
		yielder:  builtins.Runtime.Yielder(),
	}
}

type Evaluator struct {
	Stopped       bool
	yielder       Yielder // Yield to give JavaScript/browser events a chance to run.
	builtins      Builtins
	eventHandlers map[string]*parser.EventHandlerStmt

	scope  *scope // Current top of scope stack
	global *scope // Global scope
}

type Event struct {
	Name   string
	Params []any
}

func (e *Evaluator) Run(input string) error {
	builtins := e.builtins.ParserBuiltins()
	prog, err := parser.Parse(input, builtins)
	if err != nil {
		return err
	}
	if _, err := e.Eval(prog); err != nil {
		return err
	}
	return nil
}

type Yielder interface {
	Yield()
}

func (e *Evaluator) Eval(node parser.Node) (Value, error) {
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
		return &Num{Val: node.Value}, nil
	case *parser.StringLiteral:
		return &String{Val: node.Value}, nil
	case *parser.Bool:
		return e.evalBool(node), nil
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
		return e.evalBreak(node), nil
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
		return e.Eval(node.Expr)
	case *parser.TypeAssertion:
		return e.evalTypeAssertion(node)
	case *parser.FuncDeclStmt, *parser.EventHandlerStmt, *parser.EmptyStmt:
		return &None{}, nil
	}
	return nil, fmt.Errorf("%w: %v", ErrUnknownNode, node)
}

func (e *Evaluator) EventHandlerNames() []string {
	return parser.EventHandlerNames(e.eventHandlers)
}

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
		e.scope.set(param.Name, arg, param.Type())
	}
	_, err := e.Eval(eh.Body)
	return err
}

func (e *Evaluator) yield() {
	if e.yielder != nil {
		e.yielder.Yield()
	}
}

func (e *Evaluator) evalProgram(program *parser.Program) (Value, error) {
	e.eventHandlers = program.EventHandlers
	return e.evalStatments(program.Statements)
}

func (e *Evaluator) evalStatments(statements []parser.Node) (Value, error) {
	var result Value
	for _, statement := range statements {
		result, err := e.Eval(statement)
		if err != nil {
			return nil, err
		}

		if isReturn(result) || isBreak(result) { // TODO: make single: breakFlow check
			return result, nil
		}
	}
	return result, nil
}

func (e *Evaluator) evalBool(b *parser.Bool) Value {
	return &Bool{Val: b.Value}
}

func (e *Evaluator) evalDecl(decl *parser.Decl) error {
	val, err := e.Eval(decl.Value)
	if err != nil {
		return err
	}
	if decl.Type() == parser.ANY_TYPE && val.Type() != parser.ANY_TYPE {
		val = &Any{Val: val}
	}
	e.scope.set(decl.Var.Name, copyOrRef(val), decl.Type())
	return nil
}

func (e *Evaluator) evalAssignment(assignment *parser.AssignmentStmt) error {
	val, err := e.Eval(assignment.Value)
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

func (e *Evaluator) evalArrayLiteral(arr *parser.ArrayLiteral) (Value, error) {
	elements, err := e.evalExprList(arr.Elements)
	if err != nil {
		return nil, err
	}
	return &Array{Elements: &elements, T: arr.T}, nil
}

func (e *Evaluator) evalMapLiteral(m *parser.MapLiteral) (Value, error) {
	pairs := map[string]Value{}
	for key, node := range m.Pairs {
		val, err := e.Eval(node)
		if err != nil {
			return nil, err
		}
		pairs[key] = copyOrRef(val)
	}
	order := make([]string, len(m.Order))
	copy(order, m.Order)
	return &Map{Pairs: pairs, Order: &order, T: m.T}, nil
}

func (e *Evaluator) evalFunccall(funcCall *parser.FuncCall) (Value, error) {
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
	fd := funcCall.FuncDecl
	for i, param := range fd.Params {
		e.scope.set(param.Name, args[i], param.Type())
	}
	if fd.VariadicParam != nil {
		varArg := &Array{Elements: &args, T: fd.VariadicParamType}
		e.scope.set(fd.VariadicParam.Name, varArg, fd.VariadicParamType)
	}

	funcResult, err := e.Eval(fd.Body)
	if err != nil {
		return nil, err
	}
	if returnValue, ok := funcResult.(*ReturnValue); ok {
		return returnValue.Val, nil
	}
	return &None{}, nil
}

func (e *Evaluator) evalReturn(ret *parser.ReturnStmt) (Value, error) {
	if ret.Value == nil {
		return &ReturnValue{}, nil
	}
	val, err := e.Eval(ret.Value)
	if err != nil {
		return nil, err
	}
	return &ReturnValue{Val: val}, nil
}

func (e *Evaluator) evalBreak(ret *parser.BreakStmt) Value {
	return &Break{}
}

func (e *Evaluator) evalIf(i *parser.IfStmt) (Value, error) {
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
		return e.Eval(i.Else)
	}
	return &None{}, nil
}

func (e *Evaluator) evalWhile(w *parser.WhileStmt) (Value, error) {
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

func (e *Evaluator) evalFor(f *parser.ForStmt) (Value, error) {
	e.pushScope()
	defer e.popScope()
	r, err := e.newRange(f)
	if err != nil {
		return nil, err
	}
	for r.next() {
		val, err := e.Eval(f.Block)
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
	return &None{}, nil
}

func (e *Evaluator) newRange(f *parser.ForStmt) (ranger, error) {
	if r, ok := f.Range.(*parser.StepRange); ok {
		return e.newStepRange(r, f.LoopVar)
	}
	rangeVal, err := e.Eval(f.Range)
	if err != nil {
		return nil, err
	}

	switch v := rangeVal.(type) {
	case *Array:
		aRange := &arrayRange{array: v, cur: 0}
		if f.LoopVar != nil {
			aRange.loopVar = zero(f.LoopVar.Type())
			e.scope.set(f.LoopVar.Name, aRange.loopVar, f.LoopVar.Type())
		}
		return aRange, nil
	case *String:
		sRange := &stringRange{str: v, cur: 0}
		if f.LoopVar != nil {
			sRange.loopVar = &String{}
			e.scope.set(f.LoopVar.Name, sRange.loopVar, f.LoopVar.Type())
		}
		return sRange, nil
	case *Map:
		order := make([]string, len(*v.Order))
		copy(order, *v.Order)
		mapRange := &mapRange{mapVal: v, cur: 0, order: order}
		if f.LoopVar != nil {
			mapRange.loopVar = &String{}
			e.scope.set(f.LoopVar.Name, mapRange.loopVar, f.LoopVar.Type())
		}
		return mapRange, nil
	}
	return nil, newErr(f.Range, ErrRangeType)
}

func (e *Evaluator) newStepRange(r *parser.StepRange, loopVar *parser.Var) (ranger, error) {
	start, err := e.numValWithDefault(r.Start, 0.0)
	if err != nil {
		return nil, err
	}
	stop, err := e.numVal(r.Stop)
	if err != nil {
		return nil, err
	}
	step, err := e.numValWithDefault(r.Step, 1.0)
	if err != nil {
		return nil, err
	}
	if step == 0 {
		return nil, newErr(r, fmt.Errorf("%w: step cannot be 0, infinite loop", ErrRangeValue))
	}

	sRange := &stepRange{
		cur:  start,
		stop: stop,
		step: step,
	}
	if loopVar != nil {
		loopVarVal := &Num{}
		e.scope.set(loopVar.Name, loopVarVal, loopVar.Type())
		sRange.loopVar = loopVarVal
	}
	return sRange, nil
}

func (e *Evaluator) numVal(n parser.Node) (float64, error) {
	v, err := e.Eval(n)
	if err != nil {
		return 0, err
	}
	numVal, ok := v.(*Num)
	if !ok {
		return 0, newErr(n, fmt.Errorf("%w: expected number, found %v", ErrType, v))
	}
	return numVal.Val, nil
}

func (e *Evaluator) numValWithDefault(n parser.Node, defaultVal float64) (float64, error) {
	if n == nil {
		return defaultVal, nil
	}
	return e.numVal(n)
}

func (e *Evaluator) evalConditionalBlock(condBlock *parser.ConditionalBlock) (Value, bool, error) {
	e.pushScope()
	defer e.popScope()
	cond, err := e.Eval(condBlock.Condition)
	if err != nil {
		return nil, false, err
	}
	boolCond, ok := cond.(*Bool)
	if !ok {
		err := fmt.Errorf("%w: conditional not a bool", ErrType)
		return nil, false, newErr(condBlock.Condition, err)
	}
	if boolCond.Val {
		val, err := e.Eval(condBlock.Block)
		return val, true, err
	}
	return nil, false, nil
}

func (e *Evaluator) evalBlockStatment(block *parser.BlockStatement) (Value, error) {
	return e.evalStatments(block.Statements)
}

func (e *Evaluator) evalVar(v *parser.Var) (Value, error) {
	if val, ok := e.scope.get(v.Name); ok {
		return val, nil
	}
	return nil, newErr(v, fmt.Errorf("%w: %s", ErrNoVarible, v.Name))
}

func (e *Evaluator) evalExprList(terms []parser.Node) ([]Value, error) {
	result := make([]Value, len(terms))

	for i, t := range terms {
		evaluated, err := e.Eval(t)
		if err != nil {
			return nil, err
		}
		result[i] = copyOrRef(evaluated)
	}

	return result, nil
}

func (e *Evaluator) evalUnaryExpr(expr *parser.UnaryExpression) (Value, error) {
	right, err := e.Eval(expr.Right)
	if err != nil {
		return nil, err
	}
	op := expr.Op
	switch right := right.(type) {
	case *Num:
		if op == parser.OP_MINUS {
			return &Num{Val: -right.Val}, nil
		}
	case *Bool:
		if op == parser.OP_BANG {
			return &Bool{Val: !right.Val}, nil
		}
	}
	return nil, newErr(expr, fmt.Errorf("%w (unary): %v", ErrOperation, expr))
}

func (e *Evaluator) evalBinaryExpr(expr *parser.BinaryExpression) (Value, error) {
	left, err := e.Eval(expr.Left)
	if err != nil {
		return nil, err
	}
	// We need to short-circuit the evaluation of expr.Right for and/or
	// operators. We start of treating "right" as "left" and only if
	// we cannot short-circuit do we evaluate expr.Right. If we do
	// short-circuit, it does not matter what "right" is.
	right := left
	if !canShortCircuit(expr.Op, left) {
		right, err = e.Eval(expr.Right)
		if err != nil {
			return nil, err
		}
	}
	op := expr.Op
	if op == parser.OP_EQ {
		return &Bool{Val: left.Equals(right)}, nil
	}
	if op == parser.OP_NOT_EQ {
		return &Bool{Val: !left.Equals(right)}, nil
	}
	var val Value
	switch l := left.(type) {
	case *Num:
		val, err = evalBinaryNumExpr(op, l, right.(*Num))
	case *String:
		val, err = evalBinaryStringExpr(op, l, right.(*String))
	case *Bool:
		val, err = evalBinaryBoolExpr(op, l, right.(*Bool))
	case *Array:
		val, err = evalBinaryArrayExpr(op, l, right.(*Array))
	default:
		err = fmt.Errorf("%w (binary): %v", ErrOperation, expr)
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func canShortCircuit(op parser.Operator, left Value) bool {
	l, ok := left.(*Bool)
	if !ok {
		return false
	}
	switch op {
	case parser.OP_AND:
		return !l.Val // short-circuit AND when left is false
	case parser.OP_OR:
		return l.Val // short-circuit OR when left is true
	}
	return false
}

func evalBinaryNumExpr(op parser.Operator, left, right *Num) (Value, error) {
	switch op {
	case parser.OP_PLUS:
		return &Num{Val: left.Val + right.Val}, nil
	case parser.OP_MINUS:
		return &Num{Val: left.Val - right.Val}, nil
	case parser.OP_ASTERISK:
		return &Num{Val: left.Val * right.Val}, nil
	case parser.OP_PERCENT:
		return &Num{Val: math.Mod(left.Val, right.Val)}, nil
	case parser.OP_SLASH:
		return &Num{Val: left.Val / right.Val}, nil
	case parser.OP_GT:
		return &Bool{Val: left.Val > right.Val}, nil
	case parser.OP_LT:
		return &Bool{Val: left.Val < right.Val}, nil
	case parser.OP_GTEQ:
		return &Bool{Val: left.Val >= right.Val}, nil
	case parser.OP_LTEQ:
		return &Bool{Val: left.Val <= right.Val}, nil
	}
	return nil, fmt.Errorf("%w (num): %v", ErrOperation, op.String())
}

func evalBinaryStringExpr(op parser.Operator, left, right *String) (Value, error) {
	switch op {
	case parser.OP_PLUS:
		return &String{Val: left.Val + right.Val}, nil
	case parser.OP_GT:
		return &Bool{left.Val > right.Val}, nil
	case parser.OP_LT:
		return &Bool{left.Val < right.Val}, nil
	case parser.OP_GTEQ:
		return &Bool{left.Val >= right.Val}, nil
	case parser.OP_LTEQ:
		return &Bool{left.Val <= right.Val}, nil
	}
	return nil, fmt.Errorf("%w (string): %v", ErrOperation, op.String())
}

func evalBinaryBoolExpr(op parser.Operator, left, right *Bool) (Value, error) {
	switch op {
	case parser.OP_AND:
		return &Bool{Val: left.Val && right.Val}, nil
	case parser.OP_OR:
		return &Bool{Val: left.Val || right.Val}, nil
	}
	return nil, fmt.Errorf("%w (bool): %v", ErrOperation, op.String())
}

func evalBinaryArrayExpr(op parser.Operator, left, right *Array) (Value, error) {
	if op != parser.OP_PLUS {
		return nil, fmt.Errorf("%w (array): %v", ErrOperation, op.String())
	}
	result := left.Copy()
	rightElemnts := *right.Copy().Elements
	*result.Elements = append(*result.Elements, rightElemnts...)
	return result, nil
}

func (e *Evaluator) evalTarget(node parser.Node) (Value, error) {
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

func (e *Evaluator) evalIndexExpr(expr *parser.IndexExpression, forAssign bool) (Value, error) {
	left, err := e.Eval(expr.Left)
	if err != nil {
		return nil, err
	}
	index, err := e.Eval(expr.Index)
	if err != nil {
		return nil, err
	}

	var val Value
	switch l := left.(type) {
	case *Array:
		val, err = l.Index(index)
	case *String:
		val, err = l.Index(index)
	case *Map:
		strIndex, ok := index.(*String)
		if !ok {
			return nil, newErr(expr.Left, fmt.Errorf("%w: expected string for map index, found %v", ErrType, index))
		}
		if forAssign {
			l.InsertKey(strIndex.Val, expr.Type())
		}
		val, err = l.Get(strIndex.Val)
	default:
		err = fmt.Errorf("%w: expected array, string or map with index, found %v", ErrType, left.Type())
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalDotExpr(expr *parser.DotExpression, forAssign bool) (Value, error) {
	left, err := e.Eval(expr.Left)
	if err != nil {
		return nil, err
	}
	m, ok := left.(*Map)
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

func (e *Evaluator) evalSliceExpr(expr *parser.SliceExpression) (Value, error) {
	left, err := e.Eval(expr.Left)
	if err != nil {
		return nil, err
	}
	var start, end Value
	if expr.Start != nil {
		if start, err = e.Eval(expr.Start); err != nil {
			return nil, err
		}
	}
	if expr.End != nil {
		if end, err = e.Eval(expr.End); err != nil {
			return nil, err
		}
	}
	var val Value
	switch left := left.(type) {
	case *Array:
		val, err = left.Slice(start, end)
	case *String:
		val, err = left.Slice(start, end)
	default:
		err = fmt.Errorf(`%w: expected string or array before "[", found %v`, ErrType, left)
	}
	if err != nil {
		return nil, newErr(expr, err)
	}
	return val, nil
}

func (e *Evaluator) evalTypeAssertion(ta *parser.TypeAssertion) (Value, error) {
	left, err := e.Eval(ta.Left)
	if err != nil {
		return nil, err
	}
	// The parser should have already validated that the type of Left
	// is `any`, but we check anyway just in case. This may be removed
	// later.
	a, ok := left.(*Any)
	if !ok {
		return nil, newErr(ta, fmt.Errorf("%w: not an any", ErrAnyConversion))
	}
	if !a.Val.Type().Matches(ta.T) {
		return nil, newErr(ta, fmt.Errorf("%w: expected %v, found %v", ErrAnyConversion, ta.T, a.Val.Type()))
	}
	return a.Val, nil
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
