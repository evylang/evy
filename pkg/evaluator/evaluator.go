// Package evaluator evaluates a given syntax tree as created by the
// parser packages. It also exports a Run and RunWithBuiltings function
// which creates and calls a Parser.
package evaluator

import (
	"foxygo.at/evy/pkg/parser"
)

func Run(input string, printFn func(string)) {
	RunWithBuiltins(input, printFn, DefaultBuiltins(printFn))
}

func RunWithBuiltins(input string, printFn func(string), builtins Builtins) {
	p := parser.New(input, builtins.Decls())
	prog := p.Parse()
	if p.HasErrors() {
		printFn(p.MaxErrorsString(8))
		return
	}
	e := &Evaluator{print: printFn}
	e.builtins = builtins
	val := e.Eval(newScope(), prog)
	if isError(val) {
		printFn(val.String())
	}
}

type Evaluator struct {
	print    func(string)
	builtins map[string]Builtin
}

func (e *Evaluator) Eval(scope *scope, node parser.Node) Value {
	switch node := node.(type) {
	case *parser.Program:
		return e.evalProgram(scope, node)
	case *parser.Declaration:
		return e.evalDeclaration(scope, node)
	case *parser.Assignment:
		return e.evalAssignment(scope, node)
	case *parser.Var:
		return e.evalVar(scope, node)
	case *parser.NumLiteral:
		return &Num{Val: node.Value}
	case *parser.StringLiteral:
		return &String{Val: node.Value}
	case *parser.Bool:
		return e.evalBool(node)
	case *parser.ArrayLiteral:
		return e.evalArrayLiteral(scope, node)
	case *parser.MapLiteral:
		return e.evalMapLiteral(scope, node)
	case *parser.FunctionCall:
		return e.evalFunctionCall(scope, node)
	case *parser.Return:
		return e.evalReturn(scope, node)
	case *parser.Break:
		return e.evalBreak(scope, node)
	case *parser.If:
		return e.evalIf(scope, node)
	case *parser.While:
		return e.evalWhile(scope, node)
	case *parser.For:
		return e.evalFor(scope, node)
	case *parser.BlockStatement:
		return e.evalBlockStatment(scope, node)
	case *parser.UnaryExpression:
		return e.evalUnaryExpr(scope, node)
	case *parser.BinaryExpression:
		return e.evalBinaryExpr(scope, node)
	case *parser.IndexExpression:
		return e.evalIndexExpr(scope, node)
	case *parser.SliceExpression:
		return e.evalSliceExpr(scope, node)
	case *parser.DotExpression:
		return e.evalDotExpr(scope, node)
	}
	return nil // TODO: panic?
}

func (e *Evaluator) evalProgram(scope *scope, program *parser.Program) Value {
	return e.evalStatments(scope, program.Statements)
}

func (e *Evaluator) evalStatments(scope *scope, statements []parser.Node) Value {
	var result Value
	for _, statement := range statements {
		result = e.Eval(scope, statement)
		if isError(result) || isReturn(result) || isBreak(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalBool(b *parser.Bool) Value {
	return &Bool{Val: b.Value}
}

func (e *Evaluator) evalDeclaration(scope *scope, decl *parser.Declaration) Value {
	val := e.Eval(scope, decl.Value)
	if isError(val) {
		return val
	}
	if decl.Type() == parser.ANY_TYPE && val.Type() != ANY {
		val = &Any{Val: val}
	}
	scope.set(decl.Var.Name, val)
	return nil
}

func (e *Evaluator) evalAssignment(scope *scope, assignment *parser.Assignment) Value {
	val := e.Eval(scope, assignment.Value)
	if isError(val) {
		return val
	}
	target := e.Eval(scope, assignment.Target)
	if isError(target) {
		return target
	}
	target.Set(val)
	return nil
}

func (e *Evaluator) evalArrayLiteral(scope *scope, arr *parser.ArrayLiteral) Value {
	elements := e.evalExprList(scope, arr.Elements)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &Array{Elements: &elements}
}

func (e *Evaluator) evalMapLiteral(scope *scope, m *parser.MapLiteral) Value {
	pairs := map[string]Value{}
	for key, node := range m.Pairs {
		val := e.Eval(scope, node)
		if isError(val) {
			return val
		}
		pairs[key] = val
	}
	order := make([]string, len(m.Order))
	copy(order, m.Order)
	return &Map{Pairs: pairs, Order: &order}
}

func (e *Evaluator) evalFunctionCall(scope *scope, funcCall *parser.FunctionCall) Value {
	args := e.evalExprList(scope, funcCall.Arguments)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	builtin, ok := e.builtins[funcCall.Name]
	if ok {
		return builtin.Func(args)
	}
	scope = innerScopeWithArgs(scope, funcCall.FuncDecl, args)
	funcResult := e.Eval(scope, funcCall.FuncDecl.Body)
	if returnValue, ok := funcResult.(*ReturnValue); ok {
		return returnValue.Val
	}
	return funcResult // error or nil
}

func innerScopeWithArgs(scope *scope, fd *parser.FuncDecl, args []Value) *scope {
	scope = newInnerScope(scope)
	for i, param := range fd.Params {
		scope.set(param.Name, args[i])
	}
	if fd.VariadicParam != nil {
		varArg := &Array{Elements: &args}
		scope.set(fd.VariadicParam.Name, varArg)
	}
	return scope
}

func (e *Evaluator) evalReturn(scope *scope, ret *parser.Return) Value {
	if ret.Value == nil {
		return &ReturnValue{}
	}
	val := e.Eval(scope, ret.Value)
	if isError(val) {
		return val
	}
	return &ReturnValue{Val: val}
}

func (e *Evaluator) evalBreak(scope *scope, ret *parser.Break) Value {
	return &Break{}
}

func (e *Evaluator) evalIf(scope *scope, i *parser.If) Value {
	val, ok := e.evalConditionalBlock(scope, i.IfBlock)
	if ok || isError(val) {
		return val
	}
	for _, elseif := range i.ElseIfBlocks {
		val, ok := e.evalConditionalBlock(scope, elseif)
		if ok || isError(val) {
			return val
		}
	}
	if i.Else != nil {
		return e.Eval(newInnerScope(scope), i.Else)
	}
	return nil
}

func (e *Evaluator) evalWhile(scope *scope, w *parser.While) Value {
	whileBlock := &w.ConditionalBlock
	val, ok := e.evalConditionalBlock(scope, whileBlock)
	for ok && !isError(val) && !isReturn(val) && !isBreak(val) {
		val, ok = e.evalConditionalBlock(scope, whileBlock)
	}
	return val
}

func (e *Evaluator) evalFor(scope *scope, f *parser.For) Value {
	scope = newInnerScope(scope)
	r, err := e.newRange(scope, f)
	if err != nil {
		return err
	}
	for r.next() {
		val := e.Eval(scope, f.Block)
		if isError(val) || isBreak(val) || isReturn(val) {
			return val
		}
	}
	return nil
}

func (e *Evaluator) newRange(scope *scope, f *parser.For) (ranger, Value) {
	if r, ok := f.Range.(*parser.StepRange); ok {
		return e.newStepRange(scope, r, f.LoopVar)
	}
	rangeVal := e.Eval(scope, f.Range)
	if isError(rangeVal) {
		return nil, rangeVal
	}

	switch v := rangeVal.(type) {
	case *Array:
		loopVar := zero(f.LoopVar.Type())
		scope.set(f.LoopVar.Name, loopVar)
		return &arrayRange{loopVar: loopVar, array: v, cur: 0}, nil
	case *String:
		loopVar := &String{}
		scope.set(f.LoopVar.Name, loopVar)
		return &stringRange{loopVar: loopVar, str: v, cur: 0}, nil
	case *Map:
		loopVar := &String{}
		scope.set(f.LoopVar.Name, loopVar)
		order := make([]string, len(*v.Order))
		copy(order, *v.Order)
		m := &mapRange{loopVar: loopVar, mapVal: v, cur: 0, order: order}
		return m, nil
	}
	return nil, newError("cannot create range for " + f.Range.String())
}

func (e *Evaluator) newStepRange(scope *scope, r *parser.StepRange, loopVar *parser.Var) (ranger, Value) {
	start, errValue := e.numValWithDefault(scope, r.Start, 0.0)
	if errValue != nil {
		return nil, errValue
	}
	stop, errValue := e.numVal(scope, r.Stop)
	if errValue != nil {
		return nil, errValue
	}
	step, errValue := e.numValWithDefault(scope, r.Step, 1.0)
	if errValue != nil {
		return nil, errValue
	}
	if step == 0 {
		return nil, newError("step cannot by 0, infinite loop")
	}
	loopVarVal := &Num{}
	scope.set(loopVar.Name, loopVarVal)

	ranger := &stepRange{
		loopVar: loopVarVal,
		cur:     start,
		stop:    stop,
		step:    step,
	}
	return ranger, nil
}

func (e *Evaluator) numVal(scope *scope, n parser.Node) (float64, Value) {
	v := e.Eval(scope, n)
	if isError(v) {
		return 0, v
	}
	numVal, ok := v.(*Num)
	if !ok {
		return 0, newError("expected number, found " + v.String())
	}
	return numVal.Val, nil
}

func (e *Evaluator) numValWithDefault(scope *scope, n parser.Node, defaultVal float64) (float64, Value) {
	if n == nil {
		return defaultVal, nil
	}
	return e.numVal(scope, n)
}

func (e *Evaluator) evalConditionalBlock(scope *scope, condBlock *parser.ConditionalBlock) (Value, bool) {
	scope = newInnerScope(scope)
	cond := e.Eval(scope, condBlock.Condition)
	if isError(cond) {
		return cond, false
	}
	boolCond, ok := cond.(*Bool)
	if !ok {
		return newError("conditional not a bool"), false
	}
	if boolCond.Val {
		return e.Eval(scope, condBlock.Block), true
	}
	return nil, false
}

func (e *Evaluator) evalBlockStatment(scope *scope, block *parser.BlockStatement) Value {
	return e.evalStatments(scope, block.Statements)
}

func (e *Evaluator) evalVar(scope *scope, v *parser.Var) Value {
	if val, ok := scope.get(v.Name); ok {
		return val
	}
	return newError("cannot find variable " + v.Name)
}

func (e *Evaluator) evalExprList(scope *scope, terms []parser.Node) []Value {
	result := make([]Value, len(terms))

	for i, t := range terms {
		evaluated := e.Eval(scope, t)
		if isError(evaluated) {
			return []Value{evaluated}
		}
		result[i] = evaluated
	}

	return result
}

func (e *Evaluator) evalUnaryExpr(scope *scope, expr *parser.UnaryExpression) Value {
	right := e.Eval(scope, expr.Right)
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

func (e *Evaluator) evalBinaryExpr(scope *scope, expr *parser.BinaryExpression) Value {
	left := e.Eval(scope, expr.Left)
	if isError(left) {
		return left
	}
	right := e.Eval(scope, expr.Right)
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

func (e *Evaluator) evalIndexExpr(scope *scope, expr *parser.IndexExpression) Value {
	left := e.Eval(scope, expr.Left)
	if isError(left) {
		return left
	}
	index := e.Eval(scope, expr.Index)
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
		return l.Get(strIndex.Val)
	}
	return nil
}

func (e *Evaluator) evalDotExpr(scope *scope, expr *parser.DotExpression) Value {
	left := e.Eval(scope, expr.Left)
	if isError(left) {
		return left
	}
	m, ok := left.(*Map)
	if !ok {
		return newError("expected map before '.', found " + left.String())
	}
	return m.Get(expr.Key)
}

func (e *Evaluator) evalSliceExpr(scope *scope, expr *parser.SliceExpression) Value {
	left := e.Eval(scope, expr.Left)
	if isError(left) {
		return left
	}
	start := e.Eval(scope, expr.Start)
	if isError(start) {
		return start
	}
	end := e.Eval(scope, expr.End)
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
