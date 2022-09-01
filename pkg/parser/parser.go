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
		return strings.Join(errs, "\n") + "\n\n" + prog.String()
	}
	return prog.String()
}

type Parser struct {
	l      *lexer.Lexer
	errors []Error

	pos  int          // current position in token slice (points to current token)
	cur  *lexer.Token // current token under examination
	peek *lexer.Token // next token after current token

	tokens []*lexer.Token
	funcs  map[string]int  // all function declaration by name and index in tokens.
	vars   map[string]*Var // all declared variables with type TODO: needs scoping in block statements
}

type Error struct {
	message string
	token   *lexer.Token
}

func (e Error) String() string {
	return e.token.Location() + ": " + e.message
}

func New(input string) *Parser {
	l := lexer.New(input)
	p := &Parser{l: l, pos: -1}
	return p
}

func (p *Parser) Parse() *Program {
	p.readAllTokens()
	p.vars = map[string]*Var{}
	program := &Program{Statements: []Node{}}

	p.advance()
	for p.cur.Type != lexer.EOF {
		stmt := p.parseToplevelStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.advance()
	}

	return program
}

func (p *Parser) readAllTokens() {
	p.funcs = map[string]int{"print": -1} // builtins
	token := p.l.Next()

	for ; token.Type != lexer.EOF; token = p.l.Next() {
		p.tokens = append(p.tokens, token)
		if token.Type == lexer.FUNC { // Collect all function names
			funcIndex := len(p.tokens) - 1
			token = p.l.Next()
			p.tokens = append(p.tokens, token)
			if token.Type == lexer.IDENT {
				p.funcs[token.Literal] = funcIndex
			}
		}
	}
	p.tokens = append(p.tokens, token) // append EOF with pos
}

func (p *Parser) advance() {
	if p.cur != nil && p.cur.Type == lexer.EOF {
		return
	}
	p.pos++
	p.cur = p.tokens[p.pos]
	if p.pos+1 < len(p.tokens) {
		p.peek = p.tokens[p.pos+1]
	}
}

func (p *Parser) parseToplevelStatement() Node {
	switch p.cur.Type {
	case lexer.FUNC, lexer.ON:
		return p.parseUnimplementedBlock()
	}
	return p.parseStatement()
}

func (p *Parser) parseStatement() Node {
	switch p.cur.Type {
	case lexer.FOR, lexer.WHILE, lexer.IF:
		return p.parseUnimplementedBlock()
	case lexer.RETURN, lexer.BREAK:
		return p.parseUnimplementedLine()
	case lexer.NL:
		p.advance()
		return p.parseStatement()
	case lexer.EOF:
		return nil
	case lexer.COMMENT:
		p.advance() // comment token
		p.advance() // new line
		return p.parseStatement()
	case lexer.IDENT:
		switch p.peek.Type {
		case lexer.COLON:
			return p.parseTypedDeclaration()
		case lexer.DECLARE:
			return p.parseInferredDeclaration()
		case lexer.ASSIGN, lexer.LBRACKET, lexer.DOT:
			return p.parseUnimplementedLine()
		}
		return p.parseCallStatement()
	}
	p.errors = append(p.errors, Error{message: "unexpected token " + p.cur.String(), token: p.cur})
	return nil
}

// parseTypedDeclaration parses declarations like
// `x:num` or `y:any[]{}`
func (p *Parser) parseTypedDeclaration() Node {
	decl := &Declaration{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: p.cur.Literal, Type: &TypeNode{Name: ILLEGAL_TYPE}},
	}
	p.advance() // `:` CDECLARE OLON token
	p.advance() // type
	varType := p.parseType()
	if varType.Name == ILLEGAL_TYPE {
		p.appendErrorWithToken("bad type declaration", decl.Token)
		return decl
	}
	decl.Var.Type = varType
	decl.Value = zeroValue(varType.Name, decl.Token)
	p.vars[decl.Var.Name] = decl.Var
	return decl
}

// parseType parses `num[]{}` into `MAP ARRAY NUM` inverting the order.
func (p *Parser) parseType() *TypeNode {
	result := p.parseSubType(nil)
	if result == nil {
		return &TypeNode{Name: ILLEGAL_TYPE}
	}
	return result
}

func (p *Parser) parseSubType(parent *TypeNode) *TypeNode {
	if p.isTerminator() {
		return parent
	}
	typeName := ILLEGAL_TYPE
	switch p.cur.Type {
	case lexer.NUM, lexer.STRING, lexer.BOOL, lexer.ANY:
		// basic types must be at beginning of type declaration
		if parent == nil {
			typeName = typeFromToken(p.cur.Type)
		}
	case lexer.LBRACKET, lexer.LCURLY:
		// composite types (array, map) must not be at beginning of type declaration
		if matchParen(p.cur.Type, p.peek.Type) && parent != nil {
			typeName = typeFromToken(p.cur.Type)
		}
		p.advance()
	}

	if typeName == ILLEGAL_TYPE {
		p.advanceUntilTerminator()
		return &TypeNode{Name: ILLEGAL_TYPE}
	}
	p.advance()
	return p.parseSubType(&TypeNode{Name: typeName, Sub: parent})
}

func matchParen(t1, t2 lexer.TokenType) bool {
	return (t1 == lexer.LBRACKET && t2 == lexer.RBRACKET) ||
		(t1 == lexer.LCURLY && t2 == lexer.RCURLY) ||
		(t1 == lexer.LPAREN && t2 == lexer.RPAREN)
}

func (p *Parser) parseInferredDeclaration() Node {
	decl := &Declaration{
		Token: p.cur,
		Var:   &Var{Token: p.cur, Name: p.cur.Literal, Type: &TypeNode{Name: ILLEGAL_TYPE}},
	}
	p.advance() // `:=` DECLARE token
	p.advance()
	//TODO: expressions
	if !isLiteral(p.cur.Type) {
		p.appendError("not implemented: inferred type declaration with non-literal")
		return decl
	}
	value := p.parseLiteral()
	if value == nil {
		return decl
	}
	decl.Var.Type = &TypeNode{Name: typeFromLiteral(p.cur.Type)}
	decl.Value = value
	p.vars[decl.Var.Name] = decl.Var
	return decl
}

func isLiteral(t lexer.TokenType) bool {
	return t == lexer.NUM_LIT || t == lexer.STRING_LIT ||
		t == lexer.TRUE || t == lexer.FALSE ||
		t == lexer.LBRACKET || t == lexer.LCURLY
}

func (p *Parser) parseLiteral() Node {
	// TODO: array map
	switch p.cur.Type {
	case lexer.STRING_LIT:
		return &StringLiteral{Token: p.cur, Value: p.cur.Literal}
	case lexer.NUM_LIT:
		val, err := strconv.ParseFloat(p.cur.Literal, 64)
		if err != nil {
			p.appendError(err.Error())
			return nil
		}
		return &NumLiteral{Token: p.cur, Value: val}
	case lexer.TRUE, lexer.FALSE:
		return &Bool{Token: p.cur, Value: p.cur.Type == lexer.TRUE}
	}
	return nil
}

func (p *Parser) parseCallStatement() Node {
	funcToken := p.cur
	funcName := p.cur.Literal
	if _, ok := p.funcs[funcName]; !ok {
		msg := "unknown function '" + funcName + "'"
		p.appendError(msg)
		p.advanceUntilTerminator()
		return nil
	}
	// TODO: ensure arg types match function signature.
	args := p.parseTerms()
	return &FunctionCall{
		Name:      funcName,
		Token:     funcToken,
		Arguments: args,
	}
}

func (p *Parser) parseTerms() []*Term {
	var terms []*Term
	p.advance()
	for !p.isTerminator() {
		term := p.parseTerm()
		if term != nil {
			terms = append(terms, term)
		}
		p.advance()
	}
	return terms
}

func (p *Parser) parseTerm() *Term {
	//TODO: UNARY_OP Term; composite literals; assignable; slice; type_assertion; "(" toplevel_expr ")"
	if p.cur.Type == lexer.IDENT {
		ident, ok := p.vars[p.cur.Literal]
		if !ok {
			p.appendError("unknown identifier '" + p.cur.Literal + "'")
			return nil
		}
		return &Term{
			Token: p.cur,
			Type:  ident.Type,
			Value: ident,
		}
	}
	if !isLiteral(p.cur.Type) {
		p.appendError("invalid term")
		return nil
	}
	value := p.parseLiteral()
	if value == nil {
		return nil
	}
	return &Term{
		Token: p.cur,
		Type:  &TypeNode{Name: typeFromLiteral(p.cur.Type)},
		Value: value,
	}
}

func (p *Parser) advanceUntilTerminator() {
	for !p.isTerminator() {
		p.advance()
	}
}

func (p *Parser) isTerminator() bool {
	return p.cur.Type == lexer.NL || p.cur.Type == lexer.COMMENT || p.cur.Type == lexer.EOF
}

func (p *Parser) appendError(message string) {
	p.errors = append(p.errors, Error{message: message, token: p.cur})
}

func (p *Parser) appendErrorWithToken(message string, token *lexer.Token) {
	p.errors = append(p.errors, Error{message: message, token: token})
}

func (p *Parser) parseUnimplementedLine() Node {
	p.advanceUntilTerminator()
	return nil
}

func (p *Parser) parseUnimplementedBlock() Node {
	p.advanceUntilTerminator()
	for p.cur.Type != lexer.END && p.cur.Type != lexer.EOF {
		p.parseStatement()
		p.advance()
	}
	p.advance()
	return nil
}
