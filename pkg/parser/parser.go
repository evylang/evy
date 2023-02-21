// Package parser creates an abstract syntax tree (ast) from input
// string in parser.Run() function. The parser is also responsible for
// type analysis, unreachable code analysis, unused variable analysis
// and other semantic checks. The generated ast is syntactically and
// semantically correct and may not contain any further compile time
// errors, only potential run time errors.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

type Builtins struct {
	Funcs         map[string]*FuncDeclStmt
	EventHandlers map[string]*EventHandlerStmt
	Globals       map[string]*Var
}

func Run(input string, builtins Builtins) string {
	parser := New(input, builtins)
	prog := parser.Parse()
	if len(parser.errors) > 0 {
		errs := make([]string, len(parser.errors))
		for i, e := range parser.errors {
			errs[i] = e.String()
		}
		return MaxErrorsString(parser.Errors(), 8) + "\n\n" + prog.String()
	}
	return prog.String()
}

type Parser struct {
	errors []Error

	pos  int          // current position in token slice (points to current token)
	cur  *lexer.Token // current token under examination
	peek *lexer.Token // next token after current token

	tokens        []*lexer.Token
	builtins      Builtins
	funcs         map[string]*FuncDeclStmt     // all function declarations by name
	EventHandlers map[string]*EventHandlerStmt // all event handler declarations by name

	wssStack []bool
}

// Error is an Evy parse error.
type Error struct {
	message string
	token   *lexer.Token
}

func (e Error) String() string {
	return e.token.Location() + ": " + e.message
}

func New(input string, builtins Builtins) *Parser {
	l := lexer.New(input)
	p := &Parser{
		funcs:         map[string]*FuncDeclStmt{},
		EventHandlers: map[string]*EventHandlerStmt{},
		wssStack:      []bool{false},
		builtins:      builtins,
	}
	for name, funcDecl := range builtins.Funcs {
		p.funcs[name] = funcDecl
	}

	// Read all tokens, collect function declaration tokens by index
	// funcs temporarily holds FUNC token indices for further processing
	var funcs []int
	var token *lexer.Token
	for token = l.Next(); token.Type != lexer.EOF; token = l.Next() {
		if token.Type == lexer.ILLEGAL {
			if token.Literal == `"` {
				p.appendErrorForToken(`unterminated string, missing "`, token)
			} else {
				p.appendErrorForToken("illegal character '"+token.Literal+"'", token)
			}
			continue
		}
		p.tokens = append(p.tokens, token)
		if token.Type == lexer.FUNC { // Collect all function names
			funcs = append(funcs, len(p.tokens)-1)
		}
	}
	p.tokens = append(p.tokens, token) // append EOF with pos

	// Parse all function signatures, prior to proper parsing, to build
	// a function name and type lookup table because functions can be
	// called before declaration.
	for _, i := range funcs {
		p.advanceTo(i)
		fd := p.parseFuncDeclSignature()
		if p.builtins.Globals[fd.Name] != nil {
			// We still go on to add `fd` to the funcs map so that the
			// function can be parsed correctly even though it has an invalid name.
			p.appendErrorForToken("cannot override builtin variable '"+fd.Name+"'", fd.Token)
		}
		switch {
		case p.funcs[fd.Name] == nil:
			p.funcs[fd.Name] = fd
		case builtins.Funcs[fd.Name] == nil:
			p.appendErrorForToken("redeclaration of function '"+fd.Name+"'", fd.Token)
		case builtins.Funcs[fd.Name] != nil:
			p.appendErrorForToken("cannot override builtin function '"+fd.Name+"'", fd.Token)
		}
	}
	return p
}

func (p *Parser) Errors() []Error {
	return p.errors
}

func (p *Parser) HasErrors() bool {
	return len(p.errors) != 0
}

func (p *Parser) Parse() *Program {
	return p.parseProgram()
}

// function names matching `parsePRODUCTION` align with production names
// in grammar doc/syntax_grammar.md.
func (p *Parser) parseProgram() *Program {
	program := &Program{}
	scope := newScope(nil, program) // TODO: model scope as stack like evaluator.
	for _, global := range p.builtins.Globals {
		global.isUsed = true
		scope.set(global.Name, global)
	}
	p.advanceTo(0)
	for p.cur.TokenType() != lexer.EOF {
		var stmt Node
		switch p.cur.TokenType() {
		case lexer.FUNC:
			stmt = p.parseFunc(scope)
		case lexer.ON:
			stmt = p.parseEventHandler(scope)
		default:
			tok := p.cur
			stmt = p.parseStatement(scope)
			if stmt != nil && program.AlwaysTerminates() {
				p.appendErrorForToken("unreachable code", tok)
				stmt = nil
			}
			if alwaysTerminates(stmt) {
				program.alwaysTerminates = true
			}
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	p.validateScope(scope)
	return program
}

func (p *Parser) parseFunc(scope *scope) Node {
	p.advance()  // advance past FUNC
	tok := p.cur // function name
	funcName := p.cur.Literal

	p.advancePastNL() // advance past signature, already parsed into p.funcs earlier
	fd := p.funcs[funcName]
	scope = newScopeWithReturnType(scope, fd, fd.ReturnType)
	p.addParamsToScope(scope, fd)
	block := p.parseBlock(scope) // parse to "end"

	if tok.TokenType() != lexer.IDENT {
		return nil
	}
	if fd.Body != nil {
		p.appendError("redeclaration of function '" + funcName + "'")
		return nil
	}
	if fd.ReturnType != NONE_TYPE && !block.AlwaysTerminates() {
		p.appendError("missing return")
	}
	p.assertEnd()
	p.advancePastNL()
	fd.Body = block
	return fd
}

func (p *Parser) addParamsToScope(scope *scope, fd *FuncDeclStmt) {
	for _, param := range fd.Params {
		p.validateVarDecl(scope, param, param.Token, true /* allowUnderscore */)
		scope.set(param.Name, param)
	}
	if fd.VariadicParam != nil {
		vParam := fd.VariadicParam
		p.validateVarDecl(scope, vParam, vParam.Token, true /* allowUnderscore */)

		vParamAsArray := &Var{
			Name:  vParam.Name,
			Token: vParam.Token,
			T:     &Type{Name: ARRAY, Sub: vParam.Type()},
		}
		scope.set(vParam.Name, vParamAsArray)
	}
}

func (p *Parser) parseEventHandler(scope *scope) Node {
	e := &EventHandlerStmt{Token: p.cur}
	p.advance() // advance past ON token
	if !p.assertToken(lexer.IDENT) {
		p.advancePastNL()
		return nil
	}

	e.Name = p.cur.Literal
	switch {
	case p.EventHandlers[e.Name] != nil:
		p.appendError("redeclaration of on " + e.Name)
	case p.builtins.EventHandlers[e.Name] == nil:
		p.appendError("unknown event name " + e.Name)
	default:
		p.EventHandlers[e.Name] = e
	}
	p.advance() // advance past event name IDENT
	for !p.isAtEOL() {
		p.assertToken(lexer.IDENT)
		decl := p.parseTypedDecl()
		e.Params = append(e.Params, decl.Var)
	}
	p.advancePastNL()

	scope = newScopeWithReturnType(scope, e, NONE_TYPE) // only bare returns
	p.addEventParamsToScope(scope, e)
	e.Body = p.parseBlock(scope)
	p.assertEnd()
	p.advancePastNL()
	return e
}

func (p *Parser) addEventParamsToScope(scope *scope, e *EventHandlerStmt) {
	if len(e.Params) == 0 || p.builtins.EventHandlers[e.Name] == nil {
		return
	}
	expectedParams := p.builtins.EventHandlers[e.Name].Params
	if len(e.Params) != len(expectedParams) {
		p.appendError(fmt.Sprintf("wrong number of parameters expected %d, got %d", len(expectedParams), len(e.Params)))
	}
	for i, param := range e.Params {
		p.validateVarDecl(scope, param, param.Token, true /* allowUnderscore */)
		exptectedType := expectedParams[i].Type()
		if !param.Type().Matches(exptectedType) {
			p.appendError(fmt.Sprintf("wrong type for parameter %s, expected %s, got %s", param.Name, exptectedType, param.Type()))
		}
		scope.set(param.Name, param)
	}
}

func (p *Parser) parseStatement(scope *scope) Node {
	switch p.cur.TokenType() {
	// empty statement
	case lexer.NL, lexer.EOF, lexer.COMMENT:
		p.advancePastNL()
		return nil
	case lexer.WS:
		p.advance()
		return nil
	case lexer.IDENT:
		switch p.peek.Type {
		case lexer.ASSIGN, lexer.DOT:
			return p.parseAssignmentStatement(scope)
		case lexer.COLON:
			return p.parseTypedDeclStatement(scope)
		case lexer.DECLARE:
			return p.parseInferredDeclStatement(scope)
		}
		if p.isFuncCall(p.cur) {
			return p.parseFunCallStatement(scope)
		}
		if p.peek.Type == lexer.LBRACKET {
			return p.parseAssignmentStatement(scope)
		}
		p.appendError("unknown function '" + p.cur.Literal + "'")
		p.advancePastNL()
		return nil
	case lexer.RETURN:
		return p.parseReturnStatement(scope)
	case lexer.BREAK:
		return p.parseBreakStatement(scope)
	case lexer.FOR:
		return p.parseForStatement(scope)
	case lexer.WHILE:
		return p.parseWhileStatement(scope)
	case lexer.IF:
		return p.parseIfStatement(scope)
	}
	p.appendError("unexpected input " + p.cur.FormatDetails())
	p.advancePastNL()
	return nil
}

func (p *Parser) parseAssignmentStatement(scope *scope) Node {
	if p.isFuncCall(p.cur) {
		p.appendError("cannot assign to '" + p.cur.Literal + "' as it is a function not a variable")
		p.advancePastNL()
		return nil
	}
	tok := p.cur
	target := p.parseAssignmentTarget(scope)
	if target == nil {
		p.advancePastNL()
		return nil
	}
	p.assertToken(lexer.ASSIGN)
	p.advance()
	value := p.parseTopLevelExpr(scope)
	if value == nil {
		p.advancePastNL()
		return nil
	}
	if !target.Type().Accepts(value.Type()) {
		msg := "'" + target.String() + "' accepts values of type " + target.Type().Format() + ", found " + value.Type().Format()
		p.appendErrorForToken(msg, tok)
	}
	p.assertEOL()
	p.advancePastNL()
	return &AssignmentStmt{Token: tok, Target: target, Value: value}
}

func (p *Parser) parseAssignmentTarget(scope *scope) Node {
	tok := p.cur
	name := p.cur.Literal
	p.advance()
	if name == "_" {
		p.appendErrorForToken("assignment to '_' not allowed", tok)
		return nil
	}
	v, ok := scope.get(name)
	if !ok {
		p.appendErrorForToken("unknown variable name '"+name+"'", tok)
		return nil
	}
	v.isUsed = true
	tt := p.cur.TokenType()
	var n Node = v
	for tt == lexer.LBRACKET || tt == lexer.DOT {
		if p.cur.TokenType() == lexer.LBRACKET {
			if n.Type() == STRING_TYPE {
				p.appendErrorForToken("cannot index string on left side of '=', only on right", tok)
				return nil
			}
			n = p.parseIndexOrSliceExpr(scope, n, false)
		}
		if p.cur.TokenType() == lexer.DOT {
			n = p.parseDotExpr(n)
		}
		tt = p.cur.TokenType()
	}
	return n
}

func (p *Parser) parseFuncDeclSignature() *FuncDeclStmt {
	fd := &FuncDeclStmt{Token: p.cur, ReturnType: NONE_TYPE}
	p.advance() // advance past FUNC
	if !p.assertToken(lexer.IDENT) {
		p.advancePastNL()
		return nil
	}
	fd.Name = p.cur.Literal
	p.advance() // advance past function name IDENT
	if p.cur.TokenType() == lexer.COLON {
		p.advance() // advance past `:` of return type declaration, e.g. in `func rand:num`
		fd.ReturnType = p.parseType()
		if fd.ReturnType.Name == ILLEGAL {
			p.appendErrorForToken("invalid return type: "+p.cur.FormatDetails(), fd.Token)
		}
	}
	for !p.isAtEOL() && p.cur.TokenType() != lexer.DOT3 {
		p.assertToken(lexer.IDENT)
		decl := p.parseTypedDecl()
		fd.Params = append(fd.Params, decl.Var)
	}
	if p.cur.TokenType() == lexer.DOT3 {
		p.advance()
		if len(fd.Params) == 1 {
			fd.VariadicParam = fd.Params[0]
			fd.Params = nil
		} else {
			p.appendError("invalid variadic parameter, must be used with single type")
		}
	}
	p.assertEOL()
	p.advancePastNL()
	return fd
}

func (p *Parser) parseTypedDeclStatement(scope *scope) Node {
	decl := p.parseTypedDecl()
	if decl.Type().Name != ILLEGAL && p.validateVarDecl(scope, decl.Var, decl.Token, false /* allowUnderscore */) {
		scope.set(decl.Var.Name, decl.Var)
		p.assertEOL()
	}
	p.advancePastNL()
	return decl
}

// parseTypedDecl parses declarations like
// `x:num` or `y:any[]{}`.
func (p *Parser) parseTypedDecl() *Decl {
	p.assertToken(lexer.IDENT)
	varName := p.cur.Literal
	decl := &Decl{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: varName},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:`
	v := p.parseType()
	decl.Var.T = v
	decl.Value = zeroValue(v.Name)
	if v == ILLEGAL_TYPE {
		p.appendErrorForToken("invalid type declaration for '"+varName+"'", decl.Token)
	}
	return decl
}

func (p *Parser) validateVarDecl(scope *scope, v *Var, tok *lexer.Token, allowUnderscore bool) bool {
	if _, ok := p.builtins.Globals[v.Name]; ok {
		p.appendErrorForToken("redeclaration of builtin variable '"+v.Name+"'", tok)
		return false
	}
	if scope.inLocalScope(v.Name) { // already declared in current scope
		p.appendErrorForToken("redeclaration of '"+v.Name+"'", tok)
		return false
	}
	if _, ok := p.funcs[v.Name]; ok {
		p.appendErrorForToken("invalid declaration of '"+v.Name+"', already used as function name", tok)
		return false
	}
	if !allowUnderscore && v.Name == "_" {
		p.appendErrorForToken("declaration of anonymous variable '_' not allowed here", tok)
		return false
	}
	return true
}

func (p *Parser) parseInferredDeclStatement(scope *scope) Node {
	p.assertToken(lexer.IDENT)
	varName := p.cur.Literal
	decl := &Decl{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: varName},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:=`
	valToken := p.cur
	val := p.parseTopLevelExpr(scope)
	defer p.advancePastNL()
	if val == nil || val.Type() == nil {
		p.appendError("invalid inferred declaration for '" + varName + "'")
		return nil
	}
	if val.Type() == NONE_TYPE {
		p.appendError("invalid declaration, function '" + valToken.Literal + "' has no return value")
		return nil
	}
	decl.Var.T = val.Type().Infer() // assign ANY to sub_type to empty arrays and maps.
	if !p.validateVarDecl(scope, decl.Var, decl.Token, false /* allowUnderscore */) {
		return nil
	}
	decl.Value = val
	scope.set(varName, decl.Var)
	p.assertEOL()
	return decl
}

func (p *Parser) isFuncCall(tok *lexer.Token) bool {
	funcName := tok.Literal
	_, ok := p.funcs[funcName]
	return ok
}

func (p *Parser) parseFunCallStatement(scope *scope) Node {
	fc := p.parseFuncCall(scope)
	p.assertEOL()
	p.advancePastNL()
	return fc
}

func (p *Parser) assertArgTypes(decl *FuncDeclStmt, args []Node) {
	funcName := decl.Name
	if decl.VariadicParam != nil {
		paramType := decl.VariadicParam.Type()
		for _, arg := range args {
			argType := arg.Type()
			if !paramType.Accepts(argType) && !paramType.Matches(argType) {
				p.appendError("'" + funcName + "' takes variadic arguments of type '" + paramType.Format() + "', found '" + argType.Format() + "'")
			}
		}
		return
	}
	if len(decl.Params) != len(args) {
		p.appendError("'" + funcName + "' takes " + quantify(len(decl.Params), "argument") + ", found " + strconv.Itoa(len(args)))
		return
	}
	for i := range args {
		paramType := decl.Params[i].Type()
		argType := args[i].Type()
		if !paramType.Accepts(argType) && !paramType.Matches(argType) {
			p.appendError("'" + funcName + "' takes " + ordinalize(i+1) + " argument of type '" + paramType.Format() + "', found '" + argType.Format() + "'")
		}
	}
}

func (p *Parser) advancePastNL() {
	tt := p.cur.TokenType()
	for tt != lexer.NL && tt != lexer.EOF {
		p.advance()
		tt = p.cur.TokenType()
	}
	if tt == lexer.NL {
		p.advance()
	}
}

func (p *Parser) isAtEOL() bool {
	return isEOL(p.cur.TokenType())
}

func isEOL(tt lexer.TokenType) bool {
	return tt == lexer.NL || tt == lexer.EOF || tt == lexer.COMMENT
}

func (p *Parser) assertToken(tt lexer.TokenType) bool {
	if p.cur.TokenType() != tt {
		p.appendError("expected " + tt.FormatDetails() + ", got " + p.cur.TokenType().FormatDetails())
		return false
	}
	return true
}

func (p *Parser) assertEOL() {
	if !p.isAtEOL() {
		p.appendError("expected end of line, found " + p.cur.FormatDetails())
	}
}

func (p *Parser) assertEnd() {
	p.assertToken(lexer.END)
}

func (p *Parser) appendError(message string) {
	p.errors = append(p.errors, Error{message: message, token: p.cur})
}

func (p *Parser) appendErrorForToken(message string, token *lexer.Token) {
	p.errors = append(p.errors, Error{message: message, token: token})
}

// validateScope ensures all variables in scope have been used.
func (p *Parser) validateScope(scope *scope) {
	for _, v := range scope.vars {
		if !v.isUsed {
			p.appendErrorForToken("'"+v.Name+"' declared but not used", v.Token)
		}
	}
}

func (p *Parser) parseBlock(scope *scope) *BlockStatement {
	endTokens := map[lexer.TokenType]bool{lexer.END: true, lexer.EOF: true}
	return p.parseBlockWithEndTokens(scope, endTokens)
}

func (p *Parser) parseIfBlock(scope *scope) *BlockStatement {
	endTokens := map[lexer.TokenType]bool{lexer.END: true, lexer.EOF: true, lexer.ELSE: true}
	return p.parseBlockWithEndTokens(scope, endTokens)
}

func (p *Parser) parseBlockWithEndTokens(scope *scope, endTokens map[lexer.TokenType]bool) *BlockStatement {
	block := &BlockStatement{Token: p.cur}
	for !endTokens[p.cur.TokenType()] {
		tok := p.cur
		stmt := p.parseStatement(scope)
		if stmt == nil {
			continue
		}
		if block.AlwaysTerminates() {
			p.appendErrorForToken("unreachable code", tok)
			continue
		}
		if alwaysTerminates(stmt) {
			block.alwaysTerminates = true
		}
		block.Statements = append(block.Statements, stmt)
	}
	if len(block.Statements) == 0 {
		p.appendErrorForToken("at least one statement is required here", block.Token)
	}
	p.validateScope(scope)
	return block
}

func (p *Parser) advance() {
	p.advanceWSS()
	if p.isWSS() {
		return
	}
	p.advanceIfWS()
	if p.peek.Type == lexer.WS {
		p.peek = p.lookAt(p.pos + 2)
	}
}

func (p *Parser) advanceIfWS() {
	if p.cur.Type == lexer.WS {
		p.advanceWSS()
	}
}

func (p *Parser) advanceIfWSEOL() {
	tt := p.cur.Type
	for tt == lexer.NL || tt == lexer.COMMENT || tt == lexer.WS {
		p.advanceWSS()
		tt = p.cur.Type
	}
}

func (p *Parser) isWSS() bool {
	return p.wssStack[len(p.wssStack)-1]
}

func (p *Parser) pushWSS(wss bool) {
	p.wssStack = append(p.wssStack, wss)
}

func (p *Parser) popWSS() {
	p.wssStack = p.wssStack[:len(p.wssStack)-1]
	if !p.isWSS() && p.cur.Type == lexer.WS {
		p.advance()
	}
}

// advanceWSS advances to the next token in whitespace sensitive (wss) manner.
func (p *Parser) advanceWSS() {
	p.pos++
	p.cur = p.lookAt(p.pos)
	p.peek = p.lookAt(p.pos + 1)
}

func (p *Parser) advanceTo(pos int) {
	p.pos = pos
	p.cur = p.lookAt(pos)
	p.peek = p.lookAt(pos + 1)
	if p.peek.Type == lexer.WS {
		p.peek = p.lookAt(p.pos + 2)
	}
}

func (p *Parser) lookAt(pos int) *lexer.Token {
	if pos >= len(p.tokens) || pos < 0 {
		return p.tokens[len(p.tokens)-1] // EOF with pos
	}
	return p.tokens[pos]
}

func MaxErrorsString(errs []Error, n int) string {
	if n != -1 && len(errs) > n {
		errs = errs[:n]
	}
	return ErrorsString(errs)
}

func ErrorsString(errs []Error) string {
	errsSrings := make([]string, len(errs))
	for i, err := range errs {
		errsSrings[i] = err.String()
	}
	return strings.Join(errsSrings, "\n")
}

func (p *Parser) parseReturnStatement(scope *scope) Node {
	ret := &ReturnStmt{Token: p.cur}
	p.advance() // advance past RETURN token
	retValueToken := p.cur
	if p.isAtEOL() { // no return value
		ret.T = NONE_TYPE
	} else {
		ret.Value = p.parseTopLevelExpr(scope)
		if ret.Value == nil {
			ret.T = ILLEGAL_TYPE
		} else {
			ret.T = ret.Value.Type()
			p.assertEOL()
		}
	}
	if !scope.returnType.Accepts(ret.T) {
		msg := "expected return value of type " + scope.returnType.Format() + ", found " + ret.T.Format()
		if scope.returnType == NONE_TYPE && ret.T != NONE_TYPE {
			msg = "expected no return value, found " + ret.T.Format()
		}
		p.appendErrorForToken(msg, retValueToken)
	}
	p.advancePastNL()
	return ret
}

func (p *Parser) parseBreakStatement(scope *scope) Node {
	breakStmt := &BreakStmt{Token: p.cur}
	if !inLoop(scope) {
		p.appendError("break is not in a loop")
	}
	p.advance() // advance past BREAK token
	p.assertEOL()
	p.advancePastNL()
	return breakStmt
}

func (p *Parser) parseForStatement(scope *scope) Node {
	forNode := &ForStmt{Token: p.cur}
	scope = newScope(scope, forNode)
	p.advance() // advance past FOR token

	if p.cur.TokenType() == lexer.IDENT {
		forNode.LoopVar = &Var{Token: p.cur, Name: p.cur.Literal, T: NONE_TYPE}
		if !p.validateVarDecl(scope, forNode.LoopVar, p.cur, false /* allowUnderscore */) {
			p.advancePastNL()
			return nil
		}
		scope.set(forNode.LoopVar.Name, forNode.LoopVar)
		p.advance() // advance past loopVarName
		p.assertToken(lexer.DECLARE)
		p.advance() // advance past :=
	}
	if !p.assertToken(lexer.RANGE) {
		p.advancePastNL()
		return nil
	}
	tok := p.cur
	p.advance() // advance past range
	nodes := p.parseExprList(scope)
	if len(nodes) == 0 {
		p.appendError("range cannot be empty")
		return nil // previous error
	}
	n := nodes[0]
	t := n.Type()
	if len(nodes) > 1 && t.Name != NUM {
		p.appendError("range with more than one argument must be num, found " + t.String())
		return nil
	}
	p.assertEOL()
	switch t.Name {
	case STRING, MAP:
		if forNode.LoopVar != nil {
			forNode.LoopVar.T = STRING_TYPE
		}
		forNode.Range = n
	case ARRAY:
		if forNode.LoopVar != nil {
			forNode.LoopVar.T = t.Infer().Sub
		}
		forNode.Range = n
	case NUM:
		if forNode.LoopVar != nil {
			forNode.LoopVar.T = NUM_TYPE
		}
		forNode.Range = p.parseStepRange(nodes, tok)
	default:
		p.appendError("expected num, string, array or map after range, found " + t.Format())
	}
	p.advancePastNL()
	forNode.Block = p.parseBlock(scope)
	p.assertEnd()
	p.advancePastNL()
	return forNode
}

func (p *Parser) parseStepRange(nodes []Node, tok *lexer.Token) *StepRange {
	if len(nodes) > 3 {
		p.appendErrorForToken("range can take up to 3 num arguments, found "+strconv.Itoa(len(nodes)), tok)
		return nil
	}
	for i, n := range nodes {
		if i >= 3 {
			break
		}
		if n.Type() != NUM_TYPE {
			p.appendErrorForToken("range expects num type for "+ordinalize(i+1)+" argument, found "+n.Type().String(), tok)
			return nil
		}
	}
	switch len(nodes) {
	case 1:
		return &StepRange{Token: tok, Start: nil, Stop: nodes[0], Step: nil}
	case 2:
		return &StepRange{Token: tok, Start: nodes[0], Stop: nodes[1], Step: nil}
	case 3:
		return &StepRange{Token: tok, Start: nodes[0], Stop: nodes[1], Step: nodes[2]}
	default:
		p.appendErrorForToken("range can take up to 3 num arguments, found "+strconv.Itoa(len(nodes)), tok)
		return nil
	}
}

func (p *Parser) parseWhileStatement(scope *scope) Node {
	while := &WhileStmt{}
	while.Token = p.cur
	p.advance() // advance past WHILE token
	scope = newScope(scope, while)
	while.Condition = p.parseCondition(scope)
	p.advancePastNL()
	while.Block = p.parseBlock(scope)
	p.assertEnd()
	p.advancePastNL()
	return while
}

func inLoop(s *scope) bool {
	for ; s != nil; s = s.outer {
		switch s.block.(type) {
		case *WhileStmt, *ForStmt:
			return true
		}
	}
	return false
}

func (p *Parser) parseIfStatement(scope *scope) Node {
	ifStmt := &IfStmt{Token: p.cur}
	ifStmt.IfBlock = p.parseIfConditionalBlock(newScope(scope, ifStmt))
	// else if blocks
	for p.cur.TokenType() == lexer.ELSE && p.peek.TokenType() == lexer.IF {
		p.advance() // advance past ELSE token
		elseIfBlock := p.parseIfConditionalBlock(newScope(scope, ifStmt))
		ifStmt.ElseIfBlocks = append(ifStmt.ElseIfBlocks, elseIfBlock)
	}
	// else block
	if p.cur.TokenType() == lexer.ELSE {
		p.advance() // advance past ELSE token
		p.assertEOL()
		p.advancePastNL()
		ifStmt.Else = p.parseBlock(newScope(scope, ifStmt))
	}
	p.assertEnd()
	p.advancePastNL()
	return ifStmt
}

func (p *Parser) parseIfConditionalBlock(scope *scope) *ConditionalBlock {
	ifBlock := &ConditionalBlock{Token: p.cur}
	p.advance() // advance past IF token
	ifBlock.Condition = p.parseCondition(scope)
	p.advancePastNL()
	ifBlock.Block = p.parseIfBlock(scope)
	return ifBlock
}

func (p *Parser) parseCondition(scope *scope) Node {
	tok := p.cur
	condition := p.parseTopLevelExpr(scope)
	if condition != nil {
		p.assertEOL()
		if condition.Type() != BOOL_TYPE {
			p.appendErrorForToken("expected condition of type bool, found "+condition.Type().Format(), tok)
		}
	}
	return condition
}

// parseType parses `[]{}num` into
// `{Name: ARRAY, Sub: {Name: MAP Sub: NUM_TYPE}}`.
func (p *Parser) parseType() *Type {
	tt := p.cur.TokenType()
	p.advance()
	switch tt {
	case lexer.NUM:
		return NUM_TYPE
	case lexer.STRING:
		return STRING_TYPE
	case lexer.BOOL:
		return BOOL_TYPE
	case lexer.ANY:
		return ANY_TYPE
	case lexer.LBRACKET, lexer.LCURLY:
		tt2 := p.cur.TokenType()
		if (tt == lexer.LBRACKET && tt2 == lexer.RBRACKET) || (tt == lexer.LCURLY && tt2 == lexer.RCURLY) {
			p.advance()
			if sub := p.parseType(); sub != ILLEGAL_TYPE {
				return &Type{Name: compositeTypeName(tt), Sub: sub}
			}
		}
	}
	return ILLEGAL_TYPE
}
