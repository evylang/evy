package parser

import (
	"strconv"
	"strings"

	"foxygo.at/evy/pkg/lexer"
)

func Run(input string) string {
	parser := New(input)
	prog := parser.Parse()
	if len(parser.errors) > 0 {
		errs := make([]string, len(parser.errors))
		for i, e := range parser.errors {
			errs[i] = e.String()
		}
		return parser.errorsString() + "\n\n" + prog.String()
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
	vars   map[string]*Var      // TODO: needs scoping in block statements; // all declared variables with type
}

// Error is an Evy parse error.
type Error struct {
	message string
	token   *lexer.Token
}

func (e Error) String() string {
	return e.token.Location() + ": " + e.message
}

func New(input string) *Parser {
	return NewWithBuiltins(input, builtins())
}

func NewWithBuiltins(input string, builtins map[string]*FuncDecl) *Parser {
	l := lexer.New(input)
	p := &Parser{
		vars:  map[string]*Var{},
		funcs: builtins,
	}

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

func builtins() map[string]*FuncDecl {
	return map[string]*FuncDecl{
		"print": &FuncDecl{
			Name:          "print",
			VariadicParam: &Var{Name: "a", nType: ANY_TYPE},
			ReturnType:    NONE_TYPE,
		},
		"len": &FuncDecl{
			Name:       "len",
			Params:     []*Var{{Name: "a", nType: ANY_TYPE}},
			ReturnType: NUM_TYPE,
		},
	}
}

func (p *Parser) Parse() *Program {
	return p.parseProgram()
}

// function names matching `parsePROCUTION` align with production names
// in grammar doc/syntax_grammar.md
func (p *Parser) parseProgram() *Program {
	program := &Program{Statements: []Node{}}
	p.advanceTo(0)
	for p.cur.TokenType() != lexer.EOF {
		var stmt Node

		switch p.cur.TokenType() {
		case lexer.FUNC:
			stmt = p.parseFunc()
		case lexer.ON:
			stmt = p.parseEventHandler()
		default:
			stmt = p.parseStatement()
		}
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}

func (p *Parser) parseFunc() Node {
	p.advance()  // advance past FUNC
	tok := p.cur // function name

	p.advancePastNL() // // advance past signature, already parsed into p.funcs earlier
	block := p.parseBlock()

	if tok.TokenType() != lexer.IDENT {
		return nil
	}
	fd := p.funcs[tok.Literal]
	if fd.Body != nil {
		p.appendError("redeclaration of function '" + tok.Literal + "'")
		return nil
	}
	fd.Body = block
	return fd
}

func (p *Parser) parseEventHandler() Node {
	p.advance() // advance past ON token
	e := &EventHandler{}
	if p.assertToken(lexer.IDENT) {
		e.Name = p.cur.Literal
		p.advance() // advance past event name IDENT
		p.assertEOL()
	}
	p.advancePastNL() // advance past `on EVENT_NAME`
	e.Body = p.parseBlock()
	return e
}

func (p *Parser) parseStatement() Node {
	switch p.cur.TokenType() {
	// empty statement
	case lexer.NL, lexer.EOF, lexer.COMMENT:
		p.advancePastNL()
		return nil
	case lexer.IDENT:
		switch p.peek.Type {
		case lexer.ASSIGN, lexer.LBRACKET, lexer.DOT:
			return p.parseAssignStatement() // TODO
		case lexer.COLON:
			return p.parseTypedDeclStatement()
		case lexer.DECLARE:
			return p.parseInferredDeclStatement()
		}
		if p.isFuncCall(p.cur) {
			return p.parseFunCallStatement()
		}
		p.appendError("unknown function '" + p.cur.Literal + "'")
		p.advancePastNL()
		return nil
	case lexer.RETURN:
		return p.parseReturnStatment() // TODO
	case lexer.BREAK:
		return p.parseBreakStatment() // TODO
	case lexer.FOR:
		return p.parseForStatment() // TODO
	case lexer.WHILE:
		return p.parseWhileStatment() // TODO
	case lexer.IF:
		return p.parseIfStatment() // TODO
	}
	p.appendError("unexpected input " + p.cur.FormatDetails())
	p.advancePastNL()
	return nil
}

func (p *Parser) parseAssignStatement() Node {
	return nil
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
	paramNames := map[string]bool{}
	for !p.isAtEOL() && p.cur.TokenType() != lexer.DOT3 {
		decl := p.parseTypedDecl().(*Declaration)
		name := decl.Var.Name
		if paramNames[name] {
			p.appendError("redeclaration of parameter '" + name + "'")
		}
		paramNames[name] = true
		fd.Params = append(fd.Params, decl.Var)
	}
	if p.cur.TokenType() == lexer.DOT3 {
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

func (p *Parser) parseTypedDeclStatement() Node {
	decl := p.parseTypedDecl()
	if decl.Type().Name != ILLEGAL {
		p.assertEOL()
	}
	p.advancePastNL()
	return decl
}

// parseTypedDecl parses declarations like
// `x:num` or `y:any[]{}`
func (p *Parser) parseTypedDecl() Node {
	ident := p.cur.Literal
	decl := &Declaration{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: ident},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:`
	v := p.parseType()
	decl.Var.nType = v
	decl.Value = zeroValue(v.Name)
	if v == ILLEGAL_TYPE {
		p.appendErrorForToken("invalid type declaration for '"+ident+"'", decl.Token)
	} else {
		p.vars[ident] = decl.Var
	}
	return decl
}

func matchParen(t1, t2 lexer.TokenType) bool {
	return (t1 == lexer.LBRACKET && t2 == lexer.RBRACKET) ||
		(t1 == lexer.LCURLY && t2 == lexer.RCURLY)
}

func (p *Parser) parseInferredDeclStatement() Node {
	ident := p.cur.Literal
	decl := &Declaration{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: ident}, // , nType: &Type{Name: ILLEGAL}},
	}
	p.advance() // advance past IDENT
	p.advance() // advance past `:=`
	valToken := p.cur
	val := p.parseTopLevelExpression()
	if val == nil || val.Type() == nil || val.Type() == ILLEGAL_TYPE {
		decl.Var.nType = ILLEGAL_TYPE
		p.appendError("invalid inferred declaration for '" + ident + "'")
	} else if val.Type() == NONE_TYPE {
		decl.Var.nType = ILLEGAL_TYPE
		p.appendError("invalid declaration, function '" + valToken.Literal + "' has no return value")
	} else {
		decl.Value = val
		decl.Var.nType = val.Type()
		p.vars[ident] = decl.Var
		p.assertEOL()
	}
	p.advancePastNL()
	return decl
}

func (p *Parser) parseTopLevelExpression() Node {
	tt := p.cur.TokenType()
	if tt == lexer.IDENT && p.isFuncCall(p.cur) {
		return p.parseFuncCall()
	}
	return p.parseExpression()
}

func (p *Parser) parseExpression() Node {
	return p.parseTerm()
}

func (p *Parser) parseTerm() Node {
	//TODO: UNARY_OP Term; composite literals; assignable; slice; type_assertion; "(" toplevel_expr ")"
	tt := p.cur.TokenType()
	if tt == lexer.IDENT {
		ident := p.cur.Literal
		p.advance()
		v, ok := p.vars[ident]
		if !ok {
			p.appendError("unknown identifier '" + ident + "'")
			return nil
		}
		return v
	}
	if p.isLiteral() {
		lit := p.parseLiteral()
		if lit == nil {
			return nil
		}
		return lit
	}
	p.appendError("unexpected character " + tt.FormatDetails())
	p.advance()
	return nil

}

func (p *Parser) isFuncCall(tok *lexer.Token) bool {
	funcName := tok.Literal
	_, ok := p.funcs[funcName]
	return ok
}

func (p *Parser) parseFunCallStatement() Node {
	fc := p.parseFuncCall()
	p.assertEOL()
	p.advancePastNL()
	return fc
}

func (p *Parser) parseFuncCall() Node {
	funcToken := p.cur
	funcName := p.cur.Literal
	decl := p.funcs[funcName]
	p.advance() // advance past function name IDENT
	args := p.parseTerms()
	p.assertArgTypes(decl, args)
	return &FunctionCall{
		Name:      funcName,
		Token:     funcToken,
		Arguments: args,
		nType:     decl.ReturnType,
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

func (p *Parser) parseTerms() []Node {
	var terms []Node
	for !p.isTermsEnd() {
		term := p.parseTerm()
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
	tt := p.cur.TokenType()
	return tt == lexer.NL || tt == lexer.EOF || tt == lexer.COMMENT
}

func (p *Parser) assertToken(tt lexer.TokenType) bool {
	if p.cur.TokenType() != tt {
		p.appendError("expected '" + tt.FormatDetails() + "', got '" + p.cur.TokenType().String() + "'")
		return false
	}
	return true
}

func (p *Parser) assertEOL() bool {
	if !p.isAtEOL() {
		p.appendError("expected end of line, found " + p.cur.FormatDetails())
		return false
	}
	return true
}

func (p *Parser) appendError(message string) {
	p.errors = append(p.errors, Error{message: message, token: p.cur})
}

func (p *Parser) appendErrorForToken(message string, token *lexer.Token) {
	p.errors = append(p.errors, Error{message: message, token: token})
}

func (p *Parser) parseBlock() *BlockStatement {
	tok := p.cur
	var stmts []Node
	for p.cur.TokenType() != lexer.END && p.cur.TokenType() != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	p.advancePastNL()
	return &BlockStatement{Token: tok, Statements: stmts}
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

func (p *Parser) errorsString() string {
	errs := make([]string, len(p.errors))
	for i, err := range p.errors {
		errs[i] = err.String()
	}
	return strings.Join(errs, "\n")
}

//TODO: implemented
func (p *Parser) parseReturnStatment() Node {
	p.advancePastNL()
	return nil
}

//TODO: implemented
func (p *Parser) parseBreakStatment() Node {
	p.advancePastNL()
	return nil
}

//TODO: implemented
func (p *Parser) parseForStatment() Node {
	p.advancePastNL()
	p.parseBlock()
	return nil
}

//TODO: implemented
func (p *Parser) parseWhileStatment() Node {
	p.advancePastNL()
	p.parseBlock()
	return nil
}

//TODO: implemented
func (p *Parser) parseIfStatment() Node {
	p.advancePastNL()
	p.parseBlock()
	return nil
}
