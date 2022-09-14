package evaluator

import (
	"foxygo.at/evy/pkg/parser"
)

func Run(input string, print func(string)) {
	builtins := DefaultBuiltins(print)
	p := parser.NewWithBuiltins(input, builtins.Decls())
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
		return &Bool{Val: node.Value}
	case *parser.FunctionCall:
		return e.evalFunctionCall(scope, node)
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

func (e *Evaluator) evalDeclaration(scope *scope, decl *parser.Declaration) Value {
	val := e.Eval(scope, decl.Value)
	if isError(val) {
		return val
	}
	scope.set(decl.Var.Name, val)
	return nil
}

func (e *Evaluator) evalFunctionCall(scope *scope, funcCall *parser.FunctionCall) Value {
	args := e.evalTerms(scope, funcCall.Arguments)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	builtin, ok := e.builtins[funcCall.Name]
	if !ok {
		return newError("cannot find builtin function " + funcCall.Name)
	}
	return builtin.Func(args)
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
