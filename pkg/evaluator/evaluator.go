package evaluator

import (
	"foxygo.at/evy/pkg/parser"
)

func Run(input string, print func(string)) {
	RunWithBuiltins(input, print, DefaultBuiltins(print))
}

func RunWithBuiltins(input string, print func(string), builtins Builtins) {
	p := parser.New(input, builtins.Decls())
	prog := p.Parse()
	if p.HasErrors() {
		print(p.MaxErrorsString(8))
		return
	}
	e := &Evaluator{print: print}
	e.builtins = builtins
	val := e.Eval(newScope(), prog)
	if isError(val) {
		print(val.String())
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
		v := e.evalVar(scope, node)
		return v
	case *parser.Term:
		return e.evalTerm(scope, node)
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
	case *parser.If:
		return e.evalIf(scope, node)
	case *parser.BlockStatement:
		return e.evalBlockStatment(scope, node)
	}
	return nil
}

func (e *Evaluator) evalProgram(scope *scope, program *parser.Program) Value {
	var result Value
	for _, statement := range program.Statements {
		result = e.Eval(scope, statement)
		if isError(result) {
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
	args := e.evalTerms(scope, funcCall.Arguments)
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
		return nil
	}
	val := e.Eval(scope, ret.Value)
	if isError(val) {
		return val
	}
	return &ReturnValue{Val: val}
}

func (e *Evaluator) evalIf(scope *scope, i *parser.If) Value {
	val := e.evalConditionalBlock(scope, i.IfBlock)
	if val == TRUE || isError(val) {
		return val
	}
	for _, elseif := range i.ElseIfBlocks {
		val := e.evalConditionalBlock(scope, elseif)
		if val == TRUE || isError(val) {
			return val
		}
	}
	if i.Else != nil {
		return e.Eval(newInnerScope(scope), i.Else)
	}
	return FALSE
}

func (e *Evaluator) evalConditionalBlock(scope *scope, condBlock *parser.ConditionalBlock) Value {
	scope = newInnerScope(scope)
	cond := e.Eval(scope, condBlock.Condition)
	if isError(cond) {
		return cond
	}
	if cond == TRUE {
		e.Eval(scope, condBlock.Block)
		return TRUE
	}
	return FALSE
}

func (e *Evaluator) evalBlockStatment(scope *scope, block *parser.BlockStatement) Value {
	for _, statement := range block.Statements {
		result := e.Eval(scope, statement)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE || rt == ERROR {
				return result
			}
		}
	}
	return nil
}

func (e *Evaluator) evalVar(scope *scope, v *parser.Var) Value {
	if val, ok := scope.get(v.Name); ok {
		return val
	}
	return newError("cannot find variable " + v.Name)
}

func (e *Evaluator) evalTerm(scope *scope, term parser.Node) Value {
	return e.Eval(scope, term)
}

func (e *Evaluator) evalTerms(scope *scope, terms []parser.Node) []Value {
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
