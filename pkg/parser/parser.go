// Package parser creates an abstract syntax tree (ast) from input
// string in parser.Run() function. The parser is also responsible for
// type analysis, unreachable code analysis, unused variable analysis
// and other semantic checks. The generated ast is syntactically and
// semantically correct and may not contain any further compile time
// errors, only potential run time errors.
package parser

import (
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

func Run(input string, builtins map[string]*FuncDecl) string {
	parser := New(input, builtins)
	prog := parser.Parse()
	if len(parser.errors) > 0 {
		errs := make([]string, len(parser.errors))
		for i, e := range parser.errors {
			errs[i] = e.String()
		}
		return parser.MaxErrorsString(8) + "\n\n" + prog.String()
	}
	return prog.String()
}

type Parser struct {
	errors []Error

	pos  int          // current position in token slice (points to current token)
	cur  *lexer.Token // current token under examination
	peek *lexer.Token // next token after current token

	tokens []*lexer.Token
	funcs  map[string]*FuncDecl // all function declaration by name and index in tokens.
}

// Error is an Evy parse error.
type Error struct {
	message string
	token   *lexer.Token
}

func (e Error) String() string {
	return e.token.Location() + ": " + e.message
}

func New(input string, builtins map[string]*FuncDecl) *Parser {
	l := lexer.New(input)
	p := &Parser{funcs: builtins}

	// Read all tokens, collect function declaration tokens by index
	// funcs temporarily holds FUNC token indices for further processing
	var funcs []int
	var token *lexer.Token
	for token = l.Next(); token.Type != lexer.EOF; token = l.Next() {
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
		if fd != nil {
			p.funcs[fd.Name] = fd
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
	return p.parseProgram(newScope())
}

// function names matching `parsePROCUTION` align with production names
// in grammar doc/syntax_grammar.md.
func (p *Parser) parseProgram(scope *scope) *Program {
	program := &Program{}
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
			if stmt != nil && program.AlwaysReturns() {
				p.appendErrorForToken("unreachable code", tok)
				stmt = nil
			}
			if alwaysReturns(stmt) {
				program.alwaysReturns = true
			}
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}

func (p *Parser) parseFunc(scope *scope) Node {
	p.advance()  // advance past FUNC
	tok := p.cur // function name
	funcName := p.cur.Literal

	p.advancePastNL() // // advance past signature, already parsed into p.funcs earlier
	fd := p.funcs[funcName]
	scope = newInnerScopeWithReturnType(scope, fd.ReturnType)
	p.addParamsToScope(scope, fd)
	block := p.parseBlock(scope) // parse to "end"

	if tok.TokenType() != lexer.IDENT {
		return nil
	}
	if fd.Body != nil {
		p.appendError("redeclaration of function '" + funcName + "'")
		return nil
	}
	if fd.ReturnType != NONE_TYPE && !block.AlwaysReturns() {
		p.appendError("missing return")
	}
	p.assertEnd()
	p.advancePastNL()
	fd.Body = block
	return fd
}

func (p *Parser) addParamsToScope(scope *scope, fd *FuncDecl) {
	for _, param := range fd.Params {
		if scope.inLocalScope(param.Name) {
			p.appendErrorForToken("redeclaration of parameter '"+param.Name+"'", param.Token)
		}
		if _, ok := p.funcs[param.Name]; ok {
			p.appendErrorForToken("invalid declaration of parameter '"+param.Name+"', already used as function name", param.Token)
		}
		scope.set(param.Name, param)
	}
	if fd.VariadicParam != nil {
		param := fd.VariadicParam
		if _, ok := p.funcs[param.Name]; ok {
			p.appendErrorForToken("invalid declaration of parameter '"+param.Name+"', already used as function name", param.Token)
		}
		scope.set(param.Name, param)
	}
}

func (p *Parser) parseEventHandler(scope *scope) Node {
	p.advance() // advance past ON token
	e := &EventHandler{}
	if p.assertToken(lexer.IDENT) {
		e.Name = p.cur.Literal
		p.advance() // advance past event name IDENT
		p.assertEOL()
	}
	p.advancePastNL() // advance past `on EVENT_NAME`
	e.Body = p.parseBlock(scope)
	p.assertEnd()
	p.advancePastNL()
	return e
}

func (p *Parser) parseStatement(scope *scope) Node {
	switch p.cur.TokenType() {
	// empty statement
	case lexer.NL, lexer.EOF, lexer.COMMENT:
		p.advancePastNL()
		return nil
	case lexer.IDENT:
		switch p.peek.Type {
		case lexer.ASSIGN, lexer.LBRACKET, lexer.DOT:
			return p.parseAssignmentStatement(scope)
		case lexer.COLON:
			return p.parseTypedDeclStatement(scope)
		case lexer.DECLARE:
			return p.parseInferredDeclStatement(scope)
		}
		if p.isFuncCall(p.cur) {
			return p.parseFunCallStatement(scope)
		}
		p.appendError("unknown function '" + p.cur.Literal + "'")
		p.advancePastNL()
		return nil
	case lexer.RETURN:
		return p.parseReturnStatement(scope)
	case lexer.BREAK:
		return p.parseBreakStatement() // TODO
	case lexer.FOR:
		return p.parseForStatement(scope) // TODO
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

	target := p.parseAssignable(scope)
	tok := p.cur
	if target == nil {
		p.advancePastNL()
		return nil
	}
	p.assertToken(lexer.ASSIGN)
	p.advance()
	value := p.parseTopLevelExpression(scope)
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
	return &Assignment{Token: tok, Target: target, Value: value}
}

func (p *Parser) parseAssignable(scope *scope) Node {
	name := p.cur.Literal
	p.advance()
	v, ok := scope.get(name)
	if !ok {
		p.appendError("unknown variable name '" + name + "'")
		return nil
	}
	return v
}

func (p *Parser) parseFuncDeclSignature() *FuncDecl {
	fd := &FuncDecl{Token: p.cur, ReturnType: NONE_TYPE}
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
	if decl.Type().Name != ILLEGAL && p.validateVar(scope, decl.Var, decl.Token) {
		scope.set(decl.Var.Name, decl.Var)
		p.assertEOL()
	}
	p.advancePastNL()
	return decl
}

// parseTypedDecl parses declarations like
// `x:num` or `y:any[]{}`.
func (p *Parser) parseTypedDecl() *Declaration {
	p.assertToken(lexer.IDENT)
	varName := p.cur.Literal
	decl := &Declaration{
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

func (p *Parser) validateVar(scope *scope, v *Var, tok *lexer.Token) bool {
	if scope.inLocalScope(v.Name) { // already declared in current scope
		p.appendErrorForToken("redeclaration of '"+v.Name+"'", tok)
		return false
	}
	if _, ok := p.funcs[v.Name]; ok {
		p.appendErrorForToken("invalid declaration of '"+v.Name+"', already used as function name", tok)
		return false
	}
	return true
}

func matchParen(t1, t2 lexer.TokenType) bool {
	return (t1 == lexer.LBRACKET && t2 == lexer.RBRACKET) ||
		(t1 == lexer.LCURLY && t2 == lexer.RCURLY)
}

func (p *Parser) parseInferredDeclStatement(scope *scope) Node {
	p.assertToken(lexer.IDENT)
	varName := p.cur.Literal
	decl := &Declaration{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: varName},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:=`
	valToken := p.cur
	val := p.parseTopLevelExpression(scope)
	defer p.advancePastNL()
	if val == nil || val.Type() == nil {
		p.appendError("invalid inferred declaration for '" + varName + "'")
		return nil
	}
	if val.Type() == NONE_TYPE {
		p.appendError("invalid declaration, function '" + valToken.Literal + "' has no return value")
		return nil
	}
	decl.Var.T = val.Type()
	if !p.validateVar(scope, decl.Var, decl.Token) {
		return nil
	}
	decl.Value = val
	scope.set(varName, decl.Var)
	p.assertEOL()
	return decl
}

func (p *Parser) parseTopLevelExpression(scope *scope) Node {
	tt := p.cur.TokenType()
	if tt == lexer.IDENT && p.isFuncCall(p.cur) {
		return p.parseFuncCall(scope)
	}
	return p.parseExpression(scope)
}

func (p *Parser) parseExpression(scope *scope) Node {
	return p.parseTerm(scope)
}

func (p *Parser) parseTerm(scope *scope) Node {
	// TODO: UNARY_OP Term; composite literals; assignable; slice; type_assertion; "(" toplevel_expr ")"
	tt := p.cur.TokenType()
	if tt == lexer.IDENT {
		if p.isFuncCall(p.cur) {
			p.appendError("function call must be parenthesized: (" + p.cur.Literal + " ...)")
			p.advance()
			return nil
		}
		return p.parseAssignable(scope)
	}
	if p.isLiteral() {
		lit := p.parseLiteral(scope)
		if lit == nil {
			return nil
		}
		return lit
	}
	p.appendError("unexpected " + tt.FormatDetails())
	p.advance()
	return nil
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

func (p *Parser) parseFuncCall(scope *scope) Node {
	funcToken := p.cur
	funcName := p.cur.Literal
	decl := p.funcs[funcName]
	p.advance() // advance past function name IDENT
	args := p.parseTerms(scope)
	p.assertArgTypes(decl, args)
	return &FunctionCall{
		Name:      funcName,
		Token:     funcToken,
		Arguments: args,
		FuncDecl:  decl,
		T:         decl.ReturnType,
	}
}

func (p *Parser) assertArgTypes(decl *FuncDecl, args []Node) {
	funcName := decl.Name
	if decl.VariadicParam != nil {
		paramType := decl.VariadicParam.Type()
		for _, arg := range args {
			if !paramType.Accepts(arg.Type()) {
				p.appendError("'" + funcName + "' takes variadic arguments of type '" + paramType.Format() + "', found '" + arg.Type().Format() + "'")
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
		if !paramType.Accepts(argType) {
			p.appendError("'" + funcName + "' takes " + ordinalize(i+1) + " argument of type '" + paramType.Format() + "', found '" + argType.Format() + "'")
		}
	}
}

func (p *Parser) parseTerms(scope *scope) []Node {
	var terms []Node
	for !p.isTermsEnd() {
		term := p.parseTerm(scope)
		if term != nil {
			terms = append(terms, term)
		}
	}
	return terms
}

func (p *Parser) isTermsEnd() bool {
	tt := p.cur.TokenType()
	return p.isAtEOL() || tt == lexer.RBRACKET || tt == lexer.RCURLY || tt == lexer.RPAREN
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
		if block.AlwaysReturns() {
			p.appendErrorForToken("unreachable code", tok)
			continue
		}
		if alwaysReturns(stmt) {
			block.alwaysReturns = true
		}
		block.Statements = append(block.Statements, stmt)
	}
	if len(block.Statements) == 0 {
		p.appendErrorForToken("at least one statement is required here", block.Token)
	}
	return block
}

func (p *Parser) advance() {
	p.pos++
	p.cur = p.lookAt(p.pos)
	p.peek = p.lookAt(p.pos + 1)
}

func (p *Parser) advanceTo(pos int) {
	p.pos = pos
	p.cur = p.lookAt(pos)
	p.peek = p.lookAt(pos + 1)
}

func (p *Parser) lookAt(pos int) *lexer.Token {
	if pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // EOF with pos
	}
	return p.tokens[pos]
}

func (p *Parser) MaxErrorsString(n int) string {
	errs := p.errors
	if n != -1 && len(errs) > n {
		errs = errs[:n]
	}
	return errString(errs)
}

func (p *Parser) ErrorsString() string {
	return errString(p.errors)
}

func errString(errs []Error) string {
	errsSrings := make([]string, len(errs))
	for i, err := range errs {
		errsSrings[i] = err.String()
	}
	return strings.Join(errsSrings, "\n")
}

func (p *Parser) parseReturnStatement(scope *scope) Node {
	ret := &Return{Token: p.cur}
	p.advance() // advance past RETURN token
	retValueToken := p.cur
	if p.isAtEOL() { // no return value
		ret.T = NONE_TYPE
	} else {
		ret.Value = p.parseTopLevelExpression(scope)
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

// TODO: implemented.
func (p *Parser) parseBreakStatement() Node {
	p.advancePastNL()
	return nil
}

// TODO: implemented.
func (p *Parser) parseForStatement(scope *scope) Node {
	scope = newInnerScope(scope)
	p.advancePastNL()
	p.parseBlock(scope)
	p.assertEnd()
	p.advancePastNL()
	return nil
}

func (p *Parser) parseWhileStatement(scope *scope) Node {
	tok := p.cur
	p.advance() // advance past WHILE token
	scope = newInnerScope(scope)
	condition := p.parseCondition(scope)
	p.advancePastNL()
	block := p.parseBlock(scope)
	p.assertEnd()
	p.advancePastNL()
	return &While{
		ConditionalBlock{
			Token:     tok,
			Condition: condition,
			Block:     block,
		},
	}
}

func (p *Parser) parseIfStatement(scope *scope) Node {
	ifStmt := &If{Token: p.cur}
	ifStmt.IfBlock = p.parseIfConditionalBlock(scope)
	// else if blocks
	for p.cur.TokenType() == lexer.ELSE && p.peek.TokenType() == lexer.IF {
		p.advance() // advance past ELSE token
		elseIfBlock := p.parseIfConditionalBlock(scope)
		ifStmt.ElseIfBlocks = append(ifStmt.ElseIfBlocks, elseIfBlock)
	}
	// else block
	if p.cur.TokenType() == lexer.ELSE {
		p.advance() // advance past ELSE token
		p.assertEOL()
		p.advancePastNL()
		ifStmt.Else = p.parseBlock(newInnerScope(scope))
	}
	p.assertEnd()
	p.advancePastNL()
	return ifStmt
}

func (p *Parser) parseIfConditionalBlock(scope *scope) *ConditionalBlock {
	tok := p.cur
	p.advance() // advance past IF token
	scope = newInnerScope(scope)
	condition := p.parseCondition(scope)
	p.advancePastNL()
	block := p.parseIfBlock(scope)
	return &ConditionalBlock{Token: tok, Condition: condition, Block: block}
}

func (p *Parser) parseCondition(scope *scope) Node {
	tok := p.cur
	condition := p.parseTopLevelExpression(scope)
	if condition != nil {
		p.assertEOL()
		if condition.Type() != BOOL_TYPE {
			p.appendErrorForToken("expected condition of type bool, found "+condition.Type().Format(), tok)
		}
	}
	return condition
}
