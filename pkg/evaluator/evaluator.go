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
	case *parser.BlockStatement:
		return e.evalBlockStatment(scope, node)
	}
	return nil
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
	if b.Value {
		return TRUE
	}
	return FALSE
}

func (e *Evaluator) evalDeclaration(scope *scope, decl *parser.Declaration) Value {
	val := e.Eval(scope, decl.Value)
	if isError(val) {
		return val
	}
	scope.set(decl.Var.Name, val)
	return nil
}

func (e *Evaluator) evalAssignment(scope *scope, assignment *parser.Assignment) Value {
	val := e.Eval(scope, assignment.Value)
	if isError(val) {
		return val
	}
	name := assignment.Target.String()
	// We need to update the variable in the scope it was defined.
	if s, ok := scope.getScope(name); ok {
		scope = s
	}
	scope.set(name, val)
	return nil
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
		varArg := &Array{Elements: args}
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

func (e *Evaluator) evalConditionalBlock(scope *scope, condBlock *parser.ConditionalBlock) (Value, bool) {
	scope = newInnerScope(scope)
	cond := e.Eval(scope, condBlock.Condition)
	if isError(cond) {
		return cond, false
	}
	if cond == TRUE {
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
