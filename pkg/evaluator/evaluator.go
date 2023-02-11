// Package evaluator evaluates a given syntax tree as created by the
// parser packages. It also exports a Run and RunWithBuiltings function
// which creates and calls a Parser.
package evaluator

import (
	"fmt"

	"foxygo.at/evy/pkg/parser"
)

func NewEvaluator(builtins Builtins) *Evaluator {
	return &Evaluator{
		print:    builtins.Print,
		builtins: builtins,
		scope:    newScope(),
	}
}

type Evaluator struct {
	Stopped       bool
	Yielder       Yielder // Yield to give JavaScript/browser events a chance to run.
	print         func(string)
	builtins      Builtins
	eventHandlers map[string]*parser.EventHandler

	scope *scope // Current top of scope stack
}

type Event struct {
	Name   string
	Params []any
}

func (e *Evaluator) Run(input string) {
	p := parser.New(input, newParserBuiltins(e.builtins))
	prog := p.Parse()
	if p.HasErrors() {
		e.print(parser.MaxErrorsString(p.Errors(), 8))
		return
	}
	e.eventHandlers = p.EventHandlers
	val := e.Eval(prog)
	if isError(val) {
		e.print(val.String())
	}
}

type Yielder interface {
	Yield()
}

var ErrStopped = newError("stopped")

func (e *Evaluator) Eval(node parser.Node) Value {
	if e.Stopped {
		return ErrStopped
	}
	e.yield()
	switch node := node.(type) {
	case *parser.Program:
		return e.evalProgram(node)
	case *parser.Declaration:
		return e.evalDeclaration(node)
	case *parser.Assignment:
		return e.evalAssignment(node)
	case *parser.Var:
		return e.evalVar(node)
	case *parser.NumLiteral:
		return &Num{Val: node.Value}
	case *parser.StringLiteral:
		return &String{Val: node.Value}
	case *parser.Bool:
		return e.evalBool(node)
	case *parser.ArrayLiteral:
		return e.evalArrayLiteral(node)
	case *parser.MapLiteral:
		return e.evalMapLiteral(node)
	case *parser.FunctionCall:
		return e.evalFunctionCall(node)
	case *parser.Return:
		return e.evalReturn(node)
	case *parser.Break:
		return e.evalBreak(node)
	case *parser.If:
		return e.evalIf(node)
	case *parser.While:
		return e.evalWhile(node)
	case *parser.For:
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
	case *parser.FuncDecl, *parser.EventHandler:
		return nil
	}
	return newError(fmt.Sprintf("internal error: unknown node type %v", node))
}

func (e *Evaluator) EventHandlerNames() []string {
	names := make([]string, 0, len(e.eventHandlers))
	for name := range e.eventHandlers {
		names = append(names, name)
	}
	return names
}

func (e *Evaluator) HandleEvent(ev Event) {
	eh := e.eventHandlers[ev.Name]
	if eh == nil {
		panic("no event handler for " + ev.Name)
	}
	e.pushScope()
	defer e.popScope()
	args := ev.Params
	if len(args) < len(eh.Params) {
		panic("not enough arguments for " + ev.Name)
	}
	for i, param := range eh.Params {
		arg := valueFromAny(param.Type(), args[i])
		e.scope.set(param.Name, arg)
	}
	result := e.Eval(eh.Body)

	if isError(result) {
		e.print(result.String())
	}
}

func (e *Evaluator) yield() {
	if e.Yielder != nil {
		e.Yielder.Yield()
	}
}

func (e *Evaluator) evalProgram(program *parser.Program) Value {
	return e.evalStatments(program.Statements)
}

func (e *Evaluator) evalStatments(statements []parser.Node) Value {
	var result Value
	for _, statement := range statements {
		result = e.Eval(statement)

		if isError(result) || isReturn(result) || isBreak(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalBool(b *parser.Bool) Value {
	return &Bool{Val: b.Value}
}

func (e *Evaluator) evalDeclaration(decl *parser.Declaration) Value {
	val := e.Eval(decl.Value)
	if isError(val) {
		return val
	}
	if decl.Type() == parser.ANY_TYPE && val.Type() != ANY {
		val = &Any{Val: val}
	}
	e.scope.set(decl.Var.Name, copyOrRef(val))
	return nil
}

func (e *Evaluator) evalAssignment(assignment *parser.Assignment) Value {
	val := e.Eval(assignment.Value)
	if isError(val) {
		return val
	}
	target := e.evalTarget(assignment.Target)
	if isError(target) {
		return target
	}
	target.Set(val)
	return nil
}

func (e *Evaluator) evalArrayLiteral(arr *parser.ArrayLiteral) Value {
	elements := e.evalExprList(arr.Elements)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &Array{Elements: &elements}
}

func (e *Evaluator) evalMapLiteral(m *parser.MapLiteral) Value {
	pairs := map[string]Value{}
	for key, node := range m.Pairs {
		val := e.Eval(node)
		if isError(val) {
			return val
		}
		pairs[key] = copyOrRef(val)
	}
	order := make([]string, len(m.Order))
	copy(order, m.Order)
	return &Map{Pairs: pairs, Order: &order}
}

func (e *Evaluator) evalFunctionCall(funcCall *parser.FunctionCall) Value {
	args := e.evalExprList(funcCall.Arguments)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	builtin, ok := e.builtins.Funcs[funcCall.Name]
	if ok {
		return builtin.Func(args)
	}
	e.pushScope()
	defer e.popScope()

	// Add func args to scope
	fd := funcCall.FuncDecl
	for i, param := range fd.Params {
		e.scope.set(param.Name, args[i])
	}
	if fd.VariadicParam != nil {
		varArg := &Array{Elements: &args}
		e.scope.set(fd.VariadicParam.Name, varArg)
	}

	funcResult := e.Eval(fd.Body)
	if returnValue, ok := funcResult.(*ReturnValue); ok {
		return returnValue.Val
	}
	return funcResult // error or nil
}

func (e *Evaluator) evalReturn(ret *parser.Return) Value {
	if ret.Value == nil {
		return &ReturnValue{}
	}
	val := e.Eval(ret.Value)
	if isError(val) {
		return val
	}
	return &ReturnValue{Val: val}
}

func (e *Evaluator) evalBreak(ret *parser.Break) Value {
	return &Break{}
}

func (e *Evaluator) evalIf(i *parser.If) Value {
	val, ok := e.evalConditionalBlock(i.IfBlock)
	if ok || isError(val) {
		return val
	}
	for _, elseif := range i.ElseIfBlocks {
		val, ok := e.evalConditionalBlock(elseif)
		if ok || isError(val) {
			return val
		}
	}
	if i.Else != nil {
		e.pushScope()
		defer e.popScope()
		return e.Eval(i.Else)
	}
	return nil
}

func (e *Evaluator) evalWhile(w *parser.While) Value {
	whileBlock := &w.ConditionalBlock
	val, ok := e.evalConditionalBlock(whileBlock)
	for ok && !isError(val) && !isReturn(val) && !isBreak(val) {
		val, ok = e.evalConditionalBlock(whileBlock)
	}
	return val
}

func (e *Evaluator) evalFor(f *parser.For) Value {
	e.pushScope()
	defer e.popScope()
	r, err := e.newRange(f)
	if err != nil {
		return err
	}
	for r.next() {
		val := e.Eval(f.Block)
		if isError(val) || isBreak(val) || isReturn(val) {
			return val
		}
	}
	return nil
}

func (e *Evaluator) newRange(f *parser.For) (ranger, Value) {
	if r, ok := f.Range.(*parser.StepRange); ok {
		return e.newStepRange(r, f.LoopVar)
	}
	rangeVal := e.Eval(f.Range)
	if isError(rangeVal) {
		return nil, rangeVal
	}

	switch v := rangeVal.(type) {
	case *Array:
		loopVar := zero(f.LoopVar.Type())
		e.scope.set(f.LoopVar.Name, loopVar)
		return &arrayRange{loopVar: loopVar, array: v, cur: 0}, nil
	case *String:
		loopVar := &String{}
		e.scope.set(f.LoopVar.Name, loopVar)
		return &stringRange{loopVar: loopVar, str: v, cur: 0}, nil
	case *Map:
		loopVar := &String{}
		e.scope.set(f.LoopVar.Name, loopVar)
		order := make([]string, len(*v.Order))
		copy(order, *v.Order)
		m := &mapRange{loopVar: loopVar, mapVal: v, cur: 0, order: order}
		return m, nil
	}
	return nil, newError("cannot create range for " + f.Range.String())
}

func (e *Evaluator) newStepRange(r *parser.StepRange, loopVar *parser.Var) (ranger, Value) {
	start, errValue := e.numValWithDefault(r.Start, 0.0)
	if errValue != nil {
		return nil, errValue
	}
	stop, errValue := e.numVal(r.Stop)
	if errValue != nil {
		return nil, errValue
	}
	step, errValue := e.numValWithDefault(r.Step, 1.0)
	if errValue != nil {
		return nil, errValue
	}
	if step == 0 {
		return nil, newError("step cannot by 0, infinite loop")
	}
	loopVarVal := &Num{}
	e.scope.set(loopVar.Name, loopVarVal)

	ranger := &stepRange{
		loopVar: loopVarVal,
		cur:     start,
		stop:    stop,
		step:    step,
	}
	return ranger, nil
}

func (e *Evaluator) numVal(n parser.Node) (float64, Value) {
	v := e.Eval(n)
	if isError(v) {
		return 0, v
	}
	numVal, ok := v.(*Num)
	if !ok {
		return 0, newError("expected number, found " + v.String())
	}
	return numVal.Val, nil
}

func (e *Evaluator) numValWithDefault(n parser.Node, defaultVal float64) (float64, Value) {
	if n == nil {
		return defaultVal, nil
	}
	return e.numVal(n)
}

func (e *Evaluator) evalConditionalBlock(condBlock *parser.ConditionalBlock) (Value, bool) {
	e.pushScope()
	defer e.popScope()
	cond := e.Eval(condBlock.Condition)
	if isError(cond) {
		return cond, false
	}
	boolCond, ok := cond.(*Bool)
	if !ok {
		return newError("conditional not a bool"), false
	}
	if boolCond.Val {
		return e.Eval(condBlock.Block), true
	}
	return nil, false
}

func (e *Evaluator) evalBlockStatment(block *parser.BlockStatement) Value {
	return e.evalStatments(block.Statements)
}

func (e *Evaluator) evalVar(v *parser.Var) Value {
	if val, ok := e.scope.get(v.Name); ok {
		return val
	}
	return newError("cannot find variable " + v.Name)
}

func (e *Evaluator) evalExprList(terms []parser.Node) []Value {
	result := make([]Value, len(terms))

	for i, t := range terms {
		evaluated := e.Eval(t)
		if isError(evaluated) {
			return []Value{evaluated}
		}
		result[i] = copyOrRef(evaluated)
	}

	return result
}

func (e *Evaluator) evalUnaryExpr(expr *parser.UnaryExpression) Value {
	right := e.Eval(expr.Right)
	if isError(right) {
		return right
	}
	op := expr.Op
	switch right := right.(type) {
	case *Num:
		if op == parser.OP_MINUS {
			return &Num{Val: -right.Val}
		}
	case *Bool:
		if op == parser.OP_BANG {
			return &Bool{Val: !right.Val}
		}
	}
	return newError("unknown unary operation: " + expr.String())
}

func (e *Evaluator) evalBinaryExpr(expr *parser.BinaryExpression) Value {
	left := e.Eval(expr.Left)
	if isError(left) {
		return left
	}
	right := e.Eval(expr.Right)
	if isError(right) {
		return right
	}
	op := expr.Op
	if op == parser.OP_EQ {
		return &Bool{Val: left.Equals(right)}
	}
	if op == parser.OP_NOT_EQ {
		return &Bool{Val: !left.Equals(right)}
	}
	switch l := left.(type) {
	case *Num:
		return evalBinaryNumExpr(op, l, right.(*Num))
	case *String:
		return evalBinaryStringExpr(op, l, right.(*String))
	case *Bool:
		return evalBinaryBoolExpr(op, l, right.(*Bool))
	case *Array:
		return evalBinaryArrayExpr(op, l, right.(*Array))
	}
	return newError("unknown binary operation: " + expr.String())
}

func evalBinaryNumExpr(op parser.Operator, left, right *Num) Value {
	switch op {
	case parser.OP_PLUS:
		return &Num{Val: left.Val + right.Val}
	case parser.OP_MINUS:
		return &Num{Val: left.Val - right.Val}
	case parser.OP_ASTERISK:
		return &Num{Val: left.Val * right.Val}
	case parser.OP_SLASH:
		return &Num{Val: left.Val / right.Val}
	case parser.OP_GT:
		return &Bool{Val: left.Val > right.Val}
	case parser.OP_LT:
		return &Bool{Val: left.Val < right.Val}
	case parser.OP_GTEQ:
		return &Bool{Val: left.Val >= right.Val}
	case parser.OP_LTEQ:
		return &Bool{Val: left.Val <= right.Val}
	}
	return newError("unknown num operation: " + op.String())
}

func evalBinaryStringExpr(op parser.Operator, left, right *String) Value {
	switch op {
	case parser.OP_PLUS:
		return &String{Val: left.Val + right.Val}
	case parser.OP_GT:
		return &Bool{left.Val > right.Val}
	case parser.OP_LT:
		return &Bool{left.Val < right.Val}
	case parser.OP_GTEQ:
		return &Bool{left.Val >= right.Val}
	case parser.OP_LTEQ:
		return &Bool{left.Val <= right.Val}
	}
	return newError("unknown string operation: " + op.String())
}

func evalBinaryBoolExpr(op parser.Operator, left, right *Bool) Value {
	switch op {
	case parser.OP_AND:
		return &Bool{Val: left.Val && right.Val}
	case parser.OP_OR:
		return &Bool{Val: left.Val || right.Val}
	}
	return newError("unknown bool operation: " + op.String())
}

func evalBinaryArrayExpr(op parser.Operator, left, right *Array) Value {
	if op != parser.OP_PLUS {
		return newError("unknown array operation: " + op.String())
	}
	result := left.Copy()
	rightElemnts := *right.Copy().Elements
	*result.Elements = append(*result.Elements, rightElemnts...)
	return result
}

func (e *Evaluator) evalTarget(node parser.Node) Value {
	switch n := node.(type) {
	case *parser.Var:
		return e.evalVar(n)
	case *parser.IndexExpression:
		return e.evalIndexExpr(n, true /* forAssign */)
	case *parser.DotExpression:
		return e.evalDotExpr(n, true /* forAssign */)
	}
	return newError("invalid assignment target " + node.String())
}

func (e *Evaluator) evalIndexExpr(expr *parser.IndexExpression, forAssign bool) Value {
	left := e.Eval(expr.Left)
	if isError(left) {
		return left
	}
	index := e.Eval(expr.Index)
	if isError(index) {
		return index
	}

	switch l := left.(type) {
	case *Array:
		return l.Index(index)
	case *String:
		return l.Index(index)
	case *Map:
		strIndex, ok := index.(*String)
		if !ok {
			return newError("expected string for map index, found " + index.String())
		}
		if forAssign {
			l.InsertKey(strIndex.Val, expr.Type())
		}
		return l.Get(strIndex.Val)
	}
	return nil
}

func (e *Evaluator) evalDotExpr(expr *parser.DotExpression, forAssign bool) Value {
	left := e.Eval(expr.Left)
	if isError(left) {
		return left
	}
	m, ok := left.(*Map)
	if !ok {
		return newError("expected map before '.', found " + left.String())
	}
	if forAssign {
		m.InsertKey(expr.Key, expr.Type())
	}
	return m.Get(expr.Key)
}

func (e *Evaluator) evalSliceExpr(expr *parser.SliceExpression) Value {
	left := e.Eval(expr.Left)
	if isError(left) {
		return left
	}
	start := e.evalIfNotNil(expr.Start)
	if isError(start) {
		return start
	}
	end := e.evalIfNotNil(expr.End)
	if isError(end) {
		return end
	}
	switch left := left.(type) {
	case *Array:
		return left.Slice(start, end)
	case *String:
		return left.Slice(start, end)
	}
	return newError("cannot slice " + left.String())
}

func (e *Evaluator) evalIfNotNil(n parser.Node) Value {
	if n == nil {
		return nil
	}
	return e.Eval(n)
}

func (e *Evaluator) pushScope() {
	e.scope = newInnerScope(e.scope)
}

func (e *Evaluator) popScope() {
	e.scope = e.scope.outer
}
