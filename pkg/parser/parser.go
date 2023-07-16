// Package parser creates an abstract syntax tree (ast) from input
// string in parser.Run() function. The parser is also responsible for
// type analysis, unreachable code analysis, unused variable analysis
// and other semantic checks. The generated ast is syntactically and
// semantically correct and may not contain any further compile time
// errors, only potential run time errors.
package parser

import (
	"errors"
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

func Parse(input string, builtins Builtins) (*Program, error) {
	parser := newParser(input, builtins)
	prog := parser.parse()
	if parser.errors != nil {
		return nil, parser.errors
	}
	return prog, nil
}

// Errors is a list of parse errors as we typically report more than a
// single parser error at a time to the end user. Errors itself also
// implements the error interfaced and can be treated like a single Error.
type Errors []*Error

func (e Errors) Error() string {
	s := make([]string, len(e))
	for i, err := range e {
		s[i] = err.Error()
	}
	return strings.Join(s, "\n")
}

func (e Errors) Truncate(length int) Errors {
	if len(e) <= length {
		return e
	}
	return e[:length]
}

func TruncateError(err error, length int) error {
	var parseErrors Errors
	if errors.As(err, &parseErrors) {
		return parseErrors.Truncate(8)
	}
	return err
}

// Error is an Evy parse error.
type Error struct {
	message string
	token   *lexer.Token
}

func (e *Error) Error() string {
	return e.token.Location() + ": " + e.message
}

type parser struct {
	errors Errors

	pos  int          // current position in token slice (points to current token)
	cur  *lexer.Token // current token under examination
	peek *lexer.Token // next token after current token

	tokens        []*lexer.Token
	builtins      Builtins
	funcs         map[string]*FuncDeclStmt     // all function declarations by name
	eventHandlers map[string]*EventHandlerStmt // all event handler declarations by name

	scope      *scope // Current top of scope stack
	wssStack   []bool
	formatting *formatting
}

func newParser(input string, builtins Builtins) *parser {
	l := lexer.New(input)
	p := &parser{
		funcs:         map[string]*FuncDeclStmt{},
		eventHandlers: map[string]*EventHandlerStmt{},
		wssStack:      []bool{false},
		builtins:      builtins,
		formatting:    newFormatting(),
	}
	for name, funcDecl := range builtins.Funcs {
		fd := *funcDecl
		p.funcs[name] = &fd
	}
	funcs := p.consumeTokens(l)
	p.parseFuncSignatures(funcs)
	return p
}

// consumeTokens reads all tokens and returns all function declaration
// tokens by index for further pre-processing.
func (p *parser) consumeTokens(l *lexer.Lexer) []int {
	var funcs []int
	var token *lexer.Token
	for token = l.Next(); token.Type != lexer.EOF; token = l.Next() {
		if token.Type == lexer.ILLEGAL {
			if token.Literal == `"` {
				p.appendErrorForToken(`unterminated string, missing "`, token)
			} else {
				msg := fmt.Sprintf("illegal character %q", token.Literal)
				p.appendErrorForToken(msg, token)
			}
			continue
		}
		p.tokens = append(p.tokens, token)
		if token.Type == lexer.FUNC { // Collect all function names
			funcs = append(funcs, len(p.tokens)-1)
		}
	}
	p.tokens = append(p.tokens, token) // append EOF with pos
	return funcs
}

// parseFuncSignatures parses all function signatures, prior to proper
// parsing. It builds a function name and type lookup table because
// functions can be called before declaration.
func (p *parser) parseFuncSignatures(funcs []int) {
	for _, i := range funcs {
		p.advanceTo(i)
		fd := p.parseFuncDeclSignature()
		if p.builtins.Globals[fd.Name] != nil {
			// We still go on to add `fd` to the funcs map so that the
			// function can be parsed correctly even though it has an invalid name.
			msg := fmt.Sprintf("cannot override builtin variable %q", fd.Name)
			p.appendErrorForToken(msg, fd.token)
		}
		if p.builtins.Funcs[fd.Name] != nil {
			msg := fmt.Sprintf("cannot override builtin function %q", fd.Name)
			p.appendErrorForToken(msg, fd.token)
		} else if p.funcs[fd.Name] != nil {
			msg := fmt.Sprintf("redeclaration of function %q", fd.Name)
			p.appendErrorForToken(msg, fd.token)
		}
		p.funcs[fd.Name] = fd // override anyway so the signature is correct for parsing the function
	}
}

func (p *parser) parse() *Program {
	return p.parseProgram()
}

// function names matching `parsePRODUCTION` align with production names
// in grammar doc/syntax_grammar.md.
func (p *parser) parseProgram() *Program {
	program := &Program{formatting: p.formatting}
	p.scope = newScope(nil, program)
	for _, global := range p.builtins.Globals {
		global.isUsed = true
		p.scope.set(global.Name, global)
	}
	p.advanceTo(0)
	for p.cur.TokenType() != lexer.EOF {
		var stmt Node
		switch p.cur.TokenType() {
		case lexer.FUNC:
			stmt = p.parseFunc()
		case lexer.ON:
			stmt = p.parseEventHandler()
		default:
			tok := p.cur
			stmt = p.parseStatement()
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
	p.validateScope()
	program.EventHandlers = p.eventHandlers
	program.CalledBuiltinFuncs = p.calledBuiltinFuncs()
	return program
}

func (p *parser) popScope() {
	p.scope = p.scope.outer
}

func (p *parser) pushScope(s *scope) {
	p.scope = s
}

func (p *parser) pushScopeWithNode(n Node) {
	p.scope = newScope(p.scope, n)
}

func (p *parser) parseFunc() Node {
	p.advance()  // advance past FUNC
	tok := p.cur // function name
	funcName := p.cur.Literal

	p.advancePastNL() // advance past signature, already parsed into p.funcs earlier
	fd := p.funcs[funcName]
	p.scope = newScopeWithReturnType(p.scope, fd, fd.ReturnType)
	defer p.popScope()
	p.addParamsToScope(fd)
	block := p.parseBlock() // parse to "end"

	if tok.TokenType() != lexer.IDENT {
		return nil
	}
	if fd.Body != nil {
		p.appendError(fmt.Sprintf("redeclaration of function %q", funcName))
		return nil
	}
	if fd.ReturnType != NONE_TYPE && !block.AlwaysTerminates() {
		p.appendError("missing return")
	}
	p.assertEnd()
	p.advance()
	p.recordComment(block)
	p.advancePastNL()
	fd.Body = block
	return fd
}

func (p *parser) addParamsToScope(fd *FuncDeclStmt) {
	for _, param := range fd.Params {
		p.validateVarDecl(param, param.token, true /* allowUnderscore */)
		p.scope.set(param.Name, param)
	}
	if fd.VariadicParam != nil {
		vParam := fd.VariadicParam
		p.validateVarDecl(vParam, vParam.token, true /* allowUnderscore */)

		vParamAsArray := &Var{
			token: vParam.token,
			Name:  vParam.Name,
			T:     &Type{Name: ARRAY, Sub: vParam.Type()},
		}
		p.scope.set(vParam.Name, vParamAsArray)
	}
}

func (p *parser) parseEventHandler() Node {
	e := &EventHandlerStmt{token: p.cur}
	p.advance() // advance past ON token
	if !p.assertToken(lexer.IDENT) {
		p.advancePastNL()
		return nil
	}

	e.Name = p.cur.Literal
	switch {
	case p.eventHandlers[e.Name] != nil:
		p.appendError("redeclaration of on " + e.Name)
	case p.builtins.EventHandlers[e.Name] == nil:
		p.appendError("unknown event name " + e.Name)
	default:
		p.eventHandlers[e.Name] = e
	}
	p.advance() // advance past event name IDENT
	for !p.isAtEOL() {
		p.assertToken(lexer.IDENT)
		decl := p.parseTypedDecl()
		e.Params = append(e.Params, decl.Var)
	}
	p.recordComment(e)
	p.advancePastNL()

	s := newScopeWithReturnType(p.scope, e, NONE_TYPE) // only bare returns
	p.pushScope(s)
	defer p.popScope()
	p.addEventParamsToScope(e)
	e.Body = p.parseBlock()
	p.assertEnd()
	p.advance()
	p.recordComment(e.Body)
	p.advancePastNL()
	return e
}

func (p *parser) addEventParamsToScope(e *EventHandlerStmt) {
	if len(e.Params) == 0 || p.builtins.EventHandlers[e.Name] == nil {
		return
	}
	expectedParams := p.builtins.EventHandlers[e.Name].Params
	if len(e.Params) != len(expectedParams) {
		p.appendError(fmt.Sprintf("wrong number of parameters expected %d, got %d", len(expectedParams), len(e.Params)))
	}
	for i, param := range e.Params {
		p.validateVarDecl(param, param.token, true /* allowUnderscore */)
		exptectedType := expectedParams[i].Type()
		if !param.Type().Matches(exptectedType) {
			p.appendError(fmt.Sprintf("wrong type for parameter %s, expected %s, got %s", param.Name, exptectedType, param.Type()))
		}
		p.scope.set(param.Name, param)
	}
}

func (p *parser) parseStatement() Node {
	switch p.cur.TokenType() {
	case lexer.WS:
		p.advance()
		return nil
	case lexer.NL, lexer.COMMENT:
		return p.parseEmptyStmt()
	case lexer.IDENT:
		switch p.peek.Type {
		case lexer.ASSIGN, lexer.DOT:
			return p.parseAssignmentStatement()
		case lexer.COLON:
			return p.parseTypedDeclStatement()
		case lexer.DECLARE:
			return p.parseInferredDeclStatement()
		}
		if p.isFuncCall(p.cur) {
			return p.parseFunCallStatement()
		}
		if p.peek.Type == lexer.LBRACKET {
			return p.parseAssignmentStatement()
		}
		p.appendError(fmt.Sprintf("unknown function %q", p.cur.Literal))
		p.advancePastNL()
		return nil
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.IF:
		return p.parseIfStatement()
	}
	p.appendError("unexpected input " + p.cur.FormatDetails())
	p.advancePastNL()
	return nil
}

func (p *parser) parseEmptyStmt() Node {
	empty := &EmptyStmt{token: p.cur}
	switch p.cur.Type {
	case lexer.NL:
		p.advance()
		return empty
	case lexer.COMMENT:
		p.recordComment(empty)
		p.advance() // COMMENT
		p.advance() // NL
		return empty
	default:
		panic("internal error: parseEmptyStmt of invalid type")
	}
}

func (p *parser) parseAssignmentStatement() Node {
	if p.isFuncCall(p.cur) {
		p.appendError(fmt.Sprintf("cannot assign to %q as it is a function not a variable", p.cur.Literal))
		p.advancePastNL()
		return nil
	}
	tok := p.cur
	target := p.parseAssignmentTarget()
	if target == nil {
		p.advancePastNL()
		return nil
	}
	p.assertToken(lexer.ASSIGN)
	p.advance()
	value := p.parseTopLevelExpr()
	if value == nil {
		p.advancePastNL()
		return nil
	}
	if !target.Type().Accepts(value.Type()) {
		msg := fmt.Sprintf("%q accepts values of type %s, found %s", target.String(), target.Type().String(), value.Type().String())
		p.appendErrorForToken(msg, tok)
	}
	p.assertEOL()
	stmt := &AssignmentStmt{token: tok, Target: target, Value: value}
	p.recordComment(stmt)
	p.advancePastNL()
	return stmt
}

func (p *parser) parseAssignmentTarget() Node {
	tok := p.cur
	name := p.cur.Literal
	p.advance()
	if name == "_" {
		p.appendErrorForToken(`assignment to "_" not allowed`, tok)
		return nil
	}
	v, ok := p.scope.get(name)
	if !ok {
		msg := fmt.Sprintf("unknown variable name %q", name)
		p.appendErrorForToken(msg, tok)
		return nil
	}
	v.isUsed = true
	tt := p.cur.TokenType()
	var n Node = v
	for n != nil && (tt == lexer.LBRACKET || tt == lexer.DOT) {
		if p.cur.TokenType() == lexer.LBRACKET {
			if n.Type() == STRING_TYPE {
				p.appendErrorForToken(`cannot index string on left side of "=", only on right`, tok)
				return nil
			}
			n = p.parseIndexOrSliceExpr(n, false)
		} else if p.cur.TokenType() == lexer.DOT {
			n = p.parseDotExpr(n)
		}
		tt = p.cur.TokenType()
	}
	return n
}

func (p *parser) parseFuncDeclSignature() *FuncDeclStmt {
	fd := &FuncDeclStmt{token: p.cur, ReturnType: NONE_TYPE}
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
			p.appendErrorForToken("invalid return type: "+p.cur.FormatDetails(), fd.token)
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
	p.recordComment(fd)
	p.advancePastNL()
	return fd
}

func (p *parser) parseTypedDeclStatement() Node {
	decl := p.parseTypedDecl()
	if decl.Type().Name != ILLEGAL && p.validateVarDecl(decl.Var, decl.token, false /* allowUnderscore */) {
		p.scope.set(decl.Var.Name, decl.Var)
		p.assertEOL()
	}
	typeDecl := &TypedDeclStmt{token: decl.token, Decl: decl}
	p.recordComment(typeDecl)
	p.advancePastNL()
	return typeDecl
}

// parseTypedDecl parses declarations like
// `x:num` or `y:any[]{}`.
func (p *parser) parseTypedDecl() *Decl {
	p.assertToken(lexer.IDENT)
	varName := p.cur.Literal
	decl := &Decl{
		token: p.cur,
		Var:   &Var{token: p.cur, Name: varName},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:`
	v := p.parseType()
	decl.Var.T = v
	decl.Value = zeroValue(v, p.cur)
	if v == ILLEGAL_TYPE {
		msg := fmt.Sprintf("invalid type declaration for %q", varName)
		p.appendErrorForToken(msg, decl.token)
	}
	return decl
}

func (p *parser) validateVarDecl(v *Var, tok *lexer.Token, allowUnderscore bool) bool {
	if _, ok := p.builtins.Globals[v.Name]; ok {
		msg := fmt.Sprintf("redeclaration of builtin variable %q", v.Name)
		p.appendErrorForToken(msg, tok)
		return false
	}
	if p.scope.inLocalScope(v.Name) { // already declared in current scope
		msg := fmt.Sprintf("redeclaration of %q", v.Name)
		p.appendErrorForToken(msg, tok)
		return false
	}
	if _, ok := p.funcs[v.Name]; ok {
		msg := fmt.Sprintf("invalid declaration of %q, already used as function name", v.Name)
		p.appendErrorForToken(msg, tok)
		return false
	}
	if !allowUnderscore && v.Name == "_" {
		p.appendErrorForToken(`declaration of anonymous variable "_" not allowed here`, tok)
		return false
	}
	return true
}

func (p *parser) parseInferredDeclStatement() Node {
	p.assertToken(lexer.IDENT)
	varName := p.cur.Literal
	decl := &Decl{
		token: p.cur,
		Var:   &Var{token: p.cur, Name: varName},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:=`
	valToken := p.cur
	val := p.parseTopLevelExpr()
	defer p.advancePastNL()
	if val == nil || val.Type() == nil {
		p.appendError(fmt.Sprintf("invalid inferred declaration for %q", varName))
		return nil
	}
	if val.Type() == NONE_TYPE {
		p.appendError(fmt.Sprintf("invalid declaration, function %q has no return value", valToken.Literal))
		return nil
	}
	decl.Var.T = val.Type().Infer() // assign ANY to sub_type to empty arrays and maps.
	if !p.validateVarDecl(decl.Var, decl.token, false /* allowUnderscore */) {
		return nil
	}
	decl.Value = val
	p.scope.set(varName, decl.Var)
	p.assertEOL()

	inferredDecl := &InferredDeclStmt{token: decl.token, Decl: decl}
	p.recordComment(inferredDecl)
	return inferredDecl
}

func (p *parser) isFuncCall(tok *lexer.Token) bool {
	funcName := tok.Literal
	_, ok := p.funcs[funcName]
	return ok
}

func (p *parser) parseFunCallStatement() Node {
	fc := p.parseFuncCall().(*FuncCall)
	p.assertEOL()
	fcs := &FuncCallStmt{token: fc.token, FuncCall: fc}
	p.recordComment(fcs)

	p.advancePastNL()
	return fcs
}

func (p *parser) assertArgTypes(decl *FuncDeclStmt, args []Node) {
	funcName := decl.Name
	if decl.VariadicParam != nil {
		paramType := decl.VariadicParam.Type()
		for _, arg := range args {
			argType := arg.Type()
			if !paramType.Accepts(argType) && !paramType.Matches(argType) {
				msg := fmt.Sprintf("%q takes variadic arguments of type %s, found %s", funcName, paramType.String(), argType.String())
				p.appendErrorForToken(msg, arg.Token())
			}
		}
		return
	}
	if len(decl.Params) != len(args) {
		tok := p.cur
		if len(args) > len(decl.Params) {
			tok = args[len(decl.Params)].Token()
		}
		msg := fmt.Sprintf("%q takes %s, found %d", funcName, quantify(len(decl.Params), "argument"), len(args))
		p.appendErrorForToken(msg, tok)
		return
	}
	for i, arg := range args {
		paramType := decl.Params[i].Type()
		argType := arg.Type()
		if !paramType.Accepts(argType) && !paramType.Matches(argType) {
			msg := fmt.Sprintf("%q takes %s argument of type %s, found %s", funcName, ordinalize(i+1), paramType.String(), argType.String())
			p.appendErrorForToken(msg, arg.Token())
		}
	}
}

func (p *parser) advancePastNL() {
	tt := p.cur.TokenType()
	for tt != lexer.NL && tt != lexer.EOF {
		p.advance()
		tt = p.cur.TokenType()
	}
	if tt == lexer.NL {
		p.advance()
	}
}

func (p *parser) isAtEOL() bool {
	return isEOL(p.cur.TokenType())
}

func isEOL(tt lexer.TokenType) bool {
	return tt == lexer.NL || tt == lexer.EOF || tt == lexer.COMMENT
}

func (p *parser) assertToken(tt lexer.TokenType) bool {
	if p.cur.TokenType() != tt {
		p.appendError("expected " + tt.FormatDetails() + ", got " + p.cur.TokenType().FormatDetails())
		return false
	}
	return true
}

func (p *parser) assertEOL() {
	if !p.isAtEOL() {
		p.appendError("expected end of line, found " + p.cur.FormatDetails())
	}
}

func (p *parser) assertEnd() {
	p.assertToken(lexer.END)
}

func (p *parser) appendError(message string) {
	p.appendErrorForToken(message, p.cur)
}

func (p *parser) appendErrorForToken(message string, token *lexer.Token) {
	if token == nil {
		err := fmt.Errorf("Token is nil for error %q\n previous errors: %w", message, p.errors)
		panic(err)
	}
	p.errors = append(p.errors, &Error{message: message, token: token})
}

// validateScope ensures all variables in scope have been used.
func (p *parser) validateScope() {
	for _, v := range p.scope.vars {
		if !v.isUsed {
			p.appendErrorForToken(fmt.Sprintf("%q declared but not used", v.Name), v.token)
		}
	}
}

func (p *parser) parseBlock() *BlockStatement {
	endTokens := map[lexer.TokenType]bool{lexer.END: true, lexer.EOF: true}
	return p.parseBlockWithEndTokens(endTokens)
}

func (p *parser) parseIfBlock() *BlockStatement {
	endTokens := map[lexer.TokenType]bool{lexer.END: true, lexer.EOF: true, lexer.ELSE: true}
	return p.parseBlockWithEndTokens(endTokens)
}

func (p *parser) parseBlockWithEndTokens(endTokens map[lexer.TokenType]bool) *BlockStatement {
	block := &BlockStatement{token: p.cur}
	for !endTokens[p.cur.TokenType()] {
		tok := p.cur
		stmt := p.parseStatement()
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
		p.appendErrorForToken("at least one statement is required here", block.token)
	}
	p.validateScope()
	return block
}

func (p *parser) advance() {
	p.advanceWSS()
	if p.isWSS() {
		return
	}
	p.advanceIfWS()
	if p.peek.Type == lexer.WS {
		p.peek = p.lookAt(p.pos + 2)
	}
}

func (p *parser) advanceIfWS() {
	if p.cur.Type == lexer.WS {
		p.advanceWSS()
	}
}

// parseMultilineWS parses multiline whitespace and comments as needed
// for formatting in Array and Map literals.
func (p *parser) parseMulitlineWS() []multilineItem {
	tt := p.cur.Type
	var multi []multilineItem
	for tt == lexer.NL || tt == lexer.COMMENT || tt == lexer.WS {
		if tt == lexer.NL {
			multi = append(multi, multilineNL)
		} else if tt == lexer.COMMENT {
			multi = append(multi, multilineComment(p.cur.Literal))
			p.advanceWSS() // advance past NL
			p.assertToken(lexer.NL)
		}
		p.advanceWSS()
		tt = p.cur.Type
	}
	return multi
}

func (p *parser) isWSS() bool {
	return p.wssStack[len(p.wssStack)-1]
}

func (p *parser) pushWSS(wss bool) {
	p.wssStack = append(p.wssStack, wss)
}

func (p *parser) popWSS() {
	p.wssStack = p.wssStack[:len(p.wssStack)-1]
	if !p.isWSS() && p.cur.Type == lexer.WS {
		p.advance()
	}
}

// advanceWSS advances to the next token in whitespace sensitive (wss) manner.
func (p *parser) advanceWSS() {
	p.pos++
	p.cur = p.lookAt(p.pos)
	p.peek = p.lookAt(p.pos + 1)
}

func (p *parser) advanceTo(pos int) {
	p.pos = pos
	p.cur = p.lookAt(pos)
	p.peek = p.lookAt(pos + 1)
	if p.peek.Type == lexer.WS {
		p.peek = p.lookAt(p.pos + 2)
	}
}

func (p *parser) lookAt(pos int) *lexer.Token {
	if pos >= len(p.tokens) || pos < 0 {
		return p.tokens[len(p.tokens)-1] // EOF with pos
	}
	return p.tokens[pos]
}

func (p *parser) parseReturnStatement() Node {
	ret := &ReturnStmt{token: p.cur}
	p.advance() // advance past RETURN token
	retValueToken := p.cur
	if p.isAtEOL() { // no return value
		ret.T = NONE_TYPE
	} else {
		ret.Value = p.parseTopLevelExpr()
		if ret.Value == nil {
			ret.T = ILLEGAL_TYPE
		} else {
			ret.T = ret.Value.Type()
			p.assertEOL()
		}
	}
	if p.scope.returnType == nil {
		p.appendErrorForToken("return statement not allowed here", retValueToken)
	} else if !p.scope.returnType.Accepts(ret.T) {
		msg := "expected return value of type " + p.scope.returnType.String() + ", found " + ret.T.String()
		if p.scope.returnType == NONE_TYPE && ret.T != NONE_TYPE {
			msg = "expected no return value, found " + ret.T.String()
		}
		p.appendErrorForToken(msg, retValueToken)
	}
	p.recordComment(ret)
	p.advancePastNL()
	return ret
}

func (p *parser) parseBreakStatement() Node {
	breakStmt := &BreakStmt{token: p.cur}
	if !inLoop(p.scope) {
		p.appendError("break is not in a loop")
	}
	p.advance() // advance past BREAK token
	p.assertEOL()
	p.recordComment(breakStmt)
	p.advancePastNL()
	return breakStmt
}

func (p *parser) parseForStatement() Node {
	forNode := &ForStmt{token: p.cur}
	p.pushScopeWithNode(forNode)
	defer p.popScope()
	p.advance() // advance past FOR token

	if p.cur.TokenType() == lexer.IDENT {
		forNode.LoopVar = &Var{token: p.cur, Name: p.cur.Literal, T: NONE_TYPE}
		if !p.validateVarDecl(forNode.LoopVar, p.cur, false /* allowUnderscore */) {
			p.advancePastNL()
			return nil
		}
		p.scope.set(forNode.LoopVar.Name, forNode.LoopVar)
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
	nodes := p.parseExprList()
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
		p.appendError("expected num, string, array or map after range, found " + t.String())
	}
	p.recordComment(forNode)
	p.advancePastNL()
	forNode.Block = p.parseBlock()
	p.assertEnd()
	p.advance()
	p.recordComment(forNode.Block)
	p.advancePastNL()
	return forNode
}

func (p *parser) parseStepRange(nodes []Node, tok *lexer.Token) *StepRange {
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
		return &StepRange{token: tok, Start: nil, Stop: nodes[0], Step: nil}
	case 2:
		return &StepRange{token: tok, Start: nodes[0], Stop: nodes[1], Step: nil}
	case 3:
		return &StepRange{token: tok, Start: nodes[0], Stop: nodes[1], Step: nodes[2]}
	default:
		p.appendErrorForToken("range can take up to 3 num arguments, found "+strconv.Itoa(len(nodes)), tok)
		return nil
	}
}

func (p *parser) parseWhileStatement() Node {
	while := &WhileStmt{}
	while.token = p.cur
	p.advance() // advance past WHILE token
	p.pushScopeWithNode(while)
	defer p.popScope()
	while.Condition = p.parseCondition()
	comment := p.curComment()
	p.advancePastNL()
	while.Block = p.parseBlock()
	p.recordCommentString(&while.ConditionalBlock, comment)
	p.assertEnd()
	p.advance()
	p.recordComment(while.ConditionalBlock.Block)
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

func (p *parser) parseIfStatement() Node {
	ifStmt := &IfStmt{token: p.cur}
	p.pushScopeWithNode(ifStmt)
	ifStmt.IfBlock = p.parseIfConditionalBlock()
	p.popScope()
	// else if blocks
	for p.cur.TokenType() == lexer.ELSE && p.peek.TokenType() == lexer.IF {
		p.advance() // advance past ELSE token
		p.pushScopeWithNode(ifStmt)
		elseIfBlock := p.parseIfConditionalBlock()
		p.popScope()
		ifStmt.ElseIfBlocks = append(ifStmt.ElseIfBlocks, elseIfBlock)
	}
	// else block
	if p.cur.TokenType() == lexer.ELSE {
		p.advance() // advance past ELSE token
		p.assertEOL()
		comment := p.curComment()
		p.advancePastNL()
		p.pushScopeWithNode(ifStmt)
		elseBlock := p.parseBlock()
		p.popScope()
		p.recordCommentString(elseBlock, comment)
		ifStmt.Else = elseBlock
	}
	p.assertEnd()
	p.advance()
	p.recordComment(ifStmt)
	p.advancePastNL()
	return ifStmt
}

func (p *parser) parseIfConditionalBlock() *ConditionalBlock {
	ifBlock := &ConditionalBlock{token: p.cur}
	p.advance() // advance past IF token
	ifBlock.Condition = p.parseCondition()
	p.recordComment(ifBlock)
	p.advancePastNL()
	ifBlock.Block = p.parseIfBlock()
	return ifBlock
}

func (p *parser) parseCondition() Node {
	tok := p.cur
	condition := p.parseTopLevelExpr()
	if condition != nil {
		p.assertEOL()
		if condition.Type() != BOOL_TYPE {
			p.appendErrorForToken("expected condition of type bool, found "+condition.Type().String(), tok)
		}
	}
	return condition
}

// parseType parses `[]{}num` into
// `{Name: ARRAY, Sub: {Name: MAP Sub: NUM_TYPE}}`.
func (p *parser) parseType() *Type {
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

func (p *parser) recordComment(n Node) {
	if p.cur.Type == lexer.COMMENT {
		p.formatting.recordComment(n, p.cur.Literal)
	}
}

func (p *parser) recordCommentString(n Node, str string) {
	if str != "" {
		p.formatting.recordComment(n, str)
	}
}

func (p *parser) curComment() string {
	if p.cur.Type == lexer.COMMENT {
		return p.cur.Literal
	}
	return ""
}

func (p *parser) calledBuiltinFuncs() []string {
	var funcs []string
	for name, funcDecl := range p.funcs {
		if _, ok := p.builtins.Funcs[name]; ok && funcDecl.isCalled {
			funcs = append(funcs, name)
		}
	}
	return funcs
}

func EventHandlerNames(eventHandlers map[string]*EventHandlerStmt) []string {
	names := make([]string, 0, len(eventHandlers))
	for name := range eventHandlers {
		names = append(names, name)
	}
	return names
}
