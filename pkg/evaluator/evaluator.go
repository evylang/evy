package evaluator

import (
	"foxygo.at/evy/pkg/parser"
)

func Run(input string, print func(string)) {
	p := parser.New(input)
	prog := p.Parse()
	e := Evaluator{print: print, builtins: builtinFuncs(print)}
	val := e.Eval(prog, NewScope())
	if isError(val) {
		print(val.Inspect())
	}
}

type Evaluator struct {
	print    func(string)
	builtins map[string]*Builtin
}

func (e *Evaluator) Eval(node parser.Node, scope *Scope) Object {
	switch node := node.(type) {
	case *parser.Program:
		return e.evalProgram(node, scope)
	case *parser.Declaration:
		return e.evalDeclaration(node, scope)
	case *parser.Var:
		return e.evalVar(node, scope)
	case *parser.NumLiteral:
		return &Num{Value: node.Value}
	case *parser.StringLiteral:
		return &String{Value: node.Value}
	case *parser.Bool:
		return boolObject(node.Value)
	case *parser.FunctionCall:
		return e.evalFunctionCall(node, scope)
	}
	return nil
}

func (e *Evaluator) evalProgram(program *parser.Program, scope *Scope) Object {
	var result Object
	for _, statement := range program.Statements {
		result = e.Eval(statement, scope)
		switch result := result.(type) {
		case *Error:
			return result
		}
	}
	return result
}

func (e *Evaluator) evalDeclaration(decl *parser.Declaration, scope *Scope) Object {
	val := e.Eval(decl.Value, scope)
	if isError(val) {
		return val
	}
	scope.Set(decl.Var.Name, val)
	return nil
}

func (e *Evaluator) evalFunctionCall(funcCall *parser.FunctionCall, scope *Scope) Object {
	args := e.evalTerms(funcCall.Arguments, scope)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
	builtin, ok := e.builtins[funcCall.Name]
	if !ok {
		return newError("cannot find builtin function " + funcCall.Name)
	}
	return builtin.Fn(args...)
}

func (e *Evaluator) evalVar(v *parser.Var, scope *Scope) Object {
	if val, ok := scope.Get(v.Name); ok {
		return val
	}
	return newError("cannot find variable " + v.Name)
}

func (e *Evaluator) evalTerms(terms []*parser.Term, scope *Scope) []Object {
	result := make([]Object, len(terms))

	for i, t := range terms {
		evaluated := e.Eval(t.Value, scope) // TODO: ensure type, better make it fully part of parsing
		if isError(evaluated) {
			return []Object{evaluated}
		}
		result[i] = evaluated
	}

	return result
}
