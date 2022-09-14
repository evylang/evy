package evaluator

import (
	"foxygo.at/evy/pkg/parser"
)

func Run(input string, print func(string)) {
	p := parser.New(input)
	prog := p.Parse()
	e := &Evaluator{print: print}
	e.builtins = newBuiltins(e)
	val := e.Eval(prog, newScope())
	if isError(val) {
		print(val.String())
	}
}

type Evaluator struct {
	print    func(string)
	builtins map[string]Builtin
}

func (e *Evaluator) Eval(node parser.Node, scope *scope) Value {
	switch node := node.(type) {
	case *parser.Program:
		return e.evalProgram(node, scope)
	case *parser.Declaration:
		return e.evalDeclaration(node, scope)
	case *parser.Var:
		v := e.evalVar(node, scope)
		return v
	case *parser.Term:
		return e.evalTerm(node, scope)
	case *parser.NumLiteral:
		return &Num{Val: node.Value}
	case *parser.StringLiteral:
		return &String{Val: node.Value}
	case *parser.Bool:
		return &Bool{Val: node.Value}
	case *parser.FunctionCall:
		return e.evalFunctionCall(node, scope)
	}
	return nil
}

func (e *Evaluator) evalProgram(program *parser.Program, scope *scope) Value {
	var result Value
	for _, statement := range program.Statements {
		result = e.Eval(statement, scope)
		if isError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalDeclaration(decl *parser.Declaration, scope *scope) Value {
	val := e.Eval(decl.Value, scope)
	if isError(val) {
		return val
	}
	scope.set(decl.Var.Name, val)
	return nil
}

func (e *Evaluator) evalFunctionCall(funcCall *parser.FunctionCall, scope *scope) Value {
	args := e.evalTerms(funcCall.Arguments, scope)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	builtin, ok := e.builtins[funcCall.Name]
	if !ok {
		return newError("cannot find builtin function " + funcCall.Name)
	}
	return builtin(args)
}

func (e *Evaluator) evalVar(v *parser.Var, scope *scope) Value {
	if val, ok := scope.get(v.Name); ok {
		return val
	}
	return newError("cannot find variable " + v.Name)
}

func (e *Evaluator) evalTerm(term parser.Node, scope *scope) Value {
	return e.Eval(term, scope)
}

func (e *Evaluator) evalTerms(terms []parser.Node, scope *scope) []Value {
	result := make([]Value, len(terms))

	for i, t := range terms {
		evaluated := e.Eval(t, scope)
		if isError(evaluated) {
			return []Value{evaluated}
		}
		result[i] = evaluated
	}

	return result
}
