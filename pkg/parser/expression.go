// This file contains a Top Down Operator Precedence or Pratt parser.
//
// It is based on Thorston Ball's monkey interpreter:
// https://github.com/juliaogris/monkey/blob/master/parser/parser.go
//
// The expression parser is implemented in receiver functions of the
// Parser struct as defined in parser.go.

package parser

import (
	"strconv"

	"foxygo.at/evy/pkg/lexer"
)

type precedence int

const (
	LOWEST      precedence = iota
	OR                     // or
	AND                    // and
	EQUALS                 // ==
	LESSGREATER            // > or <
	SUM                    // +
	PRODUCT                // *
	UNARY                  // -x  !x
	INDEX                  // array[i]
)

var precedences = map[lexer.TokenType]precedence{
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LTEQ:     LESSGREATER,
	lexer.GTEQ:     LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.OR:       OR,
	lexer.SLASH:    PRODUCT,
	lexer.ASTERISK: PRODUCT,
	lexer.PERCENT:  PRODUCT,
	lexer.AND:      AND,
	lexer.LBRACKET: INDEX,
	lexer.DOT:      INDEX,
}

func (p *parser) parseTopLevelExpr(scope *scope) Node {
	tok := p.cur
	if tok.Type == lexer.IDENT && p.funcs[tok.Literal] != nil {
		return p.parseFuncCall(scope)
	}
	return p.parseExpr(scope, LOWEST)
}

func (p *parser) parseFuncCall(scope *scope) Node {
	fc := &FuncCall{token: p.cur, Name: p.cur.Literal}
	p.advance() // advance past function name IDENT
	funcDecl := p.funcs[fc.Name]
	funcDecl.isCalled = true
	fc.FuncDecl = funcDecl
	fc.Arguments = p.parseExprList(scope)
	p.assertArgTypes(fc.FuncDecl, fc.Arguments)
	return fc
}

func (p *parser) parseExpr(scope *scope, prec precedence) Node {
	var left Node
	switch p.cur.Type {
	case lexer.IDENT:
		left = p.lookupVar(scope)
	case lexer.STRING_LIT, lexer.NUM_LIT, lexer.TRUE, lexer.FALSE, lexer.LBRACKET, lexer.LCURLY:
		left = p.parseLiteral(scope)
	case lexer.BANG, lexer.MINUS:
		left = p.parseUnaryExpr(scope)
	case lexer.LPAREN:
		left = p.parseGroupedExpr(scope)
	default:
		p.unexpectedLeftTokenError()
	}
	for left != nil && !p.isAtExprEnd() && prec < precedences[p.cur.Type] {
		tt := p.cur.Type
		switch {
		case isBinaryOp(tt):
			left = p.parseBinaryExpr(scope, left)
		case tt == lexer.LBRACKET:
			left = p.parseIndexOrSliceExpr(scope, left, true)
		case tt == lexer.DOT:
			left = p.parseDotExpr(left)
		default:
			return left
		}
	}
	return left // nil for previous error
}

func (p *parser) unexpectedLeftTokenError() {
	if p.isWSS() {
		tt := p.cur.Type
		prevTT := p.lookAt(p.pos - 1).Type
		if isBinaryOp(tt) && prevTT == lexer.WS {
			p.appendError("unexpected whitespace before " + p.cur.FormatDetails())
			return
		}
		if tt == lexer.WS && isBinaryOp(prevTT) {
			prevToken := p.lookAt(p.pos - 1)
			p.appendErrorForToken("unexpected whitespace after "+prevToken.FormatDetails(), prevToken)
			return
		}
	}
	p.appendError("unexpected " + p.cur.FormatDetails())
}

func (p *parser) isAtExprEnd() bool {
	if p.isWSS() && p.cur.Type == lexer.WS {
		return true
	}
	return p.isAtEOL()
}

func (p *parser) parseUnaryExpr(scope *scope) Node {
	tok := p.cur
	unaryExp := &UnaryExpression{token: tok, Op: op(tok)}
	p.advance() // advance past operator
	if p.lookAt(p.pos-1).Type == lexer.WS {
		p.appendErrorForToken("unexpected whitespace after '"+unaryExp.Op.String()+"'", tok)
	}
	unaryExp.Right = p.parseExpr(scope, UNARY)
	if unaryExp.Right == nil {
		return nil // previous error
	}
	p.validateUnaryType(unaryExp)
	return unaryExp
}

func (p *parser) parseBinaryExpr(scope *scope, left Node) Node {
	tok := p.cur
	expType := left.Type()
	if isComparisonOp(tok.Type) {
		expType = BOOL_TYPE
	}
	binaryExp := &BinaryExpression{token: tok, T: expType, Op: op(tok), Left: left}
	prec := precedences[tok.Type]
	p.advance() // advance past operator
	binaryExp.Right = p.parseExpr(scope, prec)
	if binaryExp.Right == nil {
		return nil // previous error
	}
	p.validateBinaryType(binaryExp)
	if p.isWSS() {
		p.formatting.recordWSS(binaryExp)
	}

	return binaryExp
}

func (p *parser) parseGroupedExpr(scope *scope) Node {
	p.pushWSS(false)
	defer p.popWSS()
	tok := p.cur
	p.advance() // advance past (
	exp := p.parseTopLevelExpr(scope)
	if !p.assertToken(lexer.RPAREN) || exp == nil {
		return nil
	}
	p.advance() // advance past )
	return &GroupExpression{token: tok, Expr: exp}
}

func (p *parser) parseIndexOrSliceExpr(scope *scope, left Node, allowSlice bool) Node {
	p.pushWSS(false)
	defer p.popWSS()
	tok := p.cur
	if p.lookAt(p.pos-1).Type == lexer.WS {
		p.appendError("unexpected whitespace before '['")
		return nil
	}
	p.advance() // advance past [
	leftType := left.Type().Name
	if leftType != ARRAY && leftType != MAP && leftType != STRING {
		p.appendErrorForToken("only array, string and map type can be indexed found "+left.Type().String(), tok)
		return nil
	}
	if p.cur.TokenType() == lexer.COLON && allowSlice { // e.g. a[:2]
		p.advance() //  advance past :
		return p.parseSlice(scope, tok, left, nil)
	}
	index := p.parseTopLevelExpr(scope)
	if index == nil {
		return nil
	}
	tt := p.cur.TokenType()
	if tt == lexer.COLON && allowSlice { // e.g. a[1:3] or a[1:]
		p.advance() // advance past :
		return p.parseSlice(scope, tok, left, index)
	}
	if !p.validateIndex(tok, leftType, index.Type()) {
		return nil
	}
	p.advance() // advance past ]
	t := left.Type().Sub
	if leftType == STRING {
		t = STRING_TYPE
	}
	return &IndexExpression{token: tok, Left: left, Index: index, T: t}
}

func (p *parser) validateIndex(tok *lexer.Token, leftType TypeName, indexType *Type) bool {
	if !p.assertToken(lexer.RBRACKET) {
		return false
	}
	if (leftType == ARRAY || leftType == STRING) && indexType != NUM_TYPE {
		p.appendErrorForToken(leftType.Name()+" index expects num, found "+indexType.String(), tok)
		return false
	}
	if leftType == MAP && indexType != STRING_TYPE {
		p.appendErrorForToken("map index expects string, found "+indexType.String(), tok)
		return false
	}
	return true
}

func (p *parser) parseSlice(scope *scope, tok *lexer.Token, left, start Node) Node {
	leftType := left.Type().Name
	if leftType != ARRAY && leftType != STRING {
		p.appendErrorForToken("only array and string be indexed sliced"+left.Type().String(), tok)
		return nil
	}

	t := left.Type()
	if p.cur.Type == lexer.RBRACKET {
		p.advance()
		return &SliceExpression{token: tok, Left: left, Start: start, End: nil, T: t}
	}
	end := p.parseTopLevelExpr(scope)
	if end == nil {
		return nil
	}
	if !p.assertToken(lexer.RBRACKET) {
		return nil
	}
	p.advance()
	return &SliceExpression{token: tok, Left: left, Start: start, End: end, T: t}
}

func (p *parser) parseDotExpr(left Node) Node {
	tok := p.cur
	if p.lookAt(p.pos-1).Type == lexer.WS {
		p.appendError("unexpected whitespace before '.'")
		return nil
	}
	if p.lookAt(p.pos+1).Type == lexer.WS {
		p.appendError("unexpected whitespace after '.'")
		return nil
	}
	p.advance() // advance past .
	leftType := left.Type().Name
	if leftType != MAP {
		p.appendErrorForToken("field access with '.' expects map type, found "+left.Type().String(), tok)
		return nil
	}
	if p.cur.TokenType() != lexer.IDENT {
		p.appendErrorForToken("expected map key, found "+p.cur.TokenType().String(), tok)
		return nil
	}
	expr := &DotExpression{token: tok, Left: left, T: left.Type().Sub, Key: p.cur.Literal}
	p.advance() // advance past key IDENT
	return expr
}

func isBinaryOp(tt lexer.TokenType) bool {
	return isComparisonOp(tt) || tt == lexer.PLUS || tt == lexer.MINUS || tt == lexer.SLASH || tt == lexer.ASTERISK || tt == lexer.PERCENT || tt == lexer.OR || tt == lexer.AND
}

func isComparisonOp(tt lexer.TokenType) bool {
	return tt == lexer.EQ || tt == lexer.NOT_EQ || tt == lexer.LT || tt == lexer.GT || tt == lexer.LTEQ || tt == lexer.GTEQ
}

func (p *parser) validateUnaryType(unaryExp *UnaryExpression) {
	tok := unaryExp.Token()
	rightType := unaryExp.Right.Type()
	switch unaryExp.Op {
	case OP_MINUS:
		if unaryExp.Right.Type() != NUM_TYPE {
			p.appendErrorForToken("'-' unary expects num type, found "+rightType.String(), tok)
		}
	case OP_BANG:
		if unaryExp.Right.Type() != BOOL_TYPE {
			p.appendErrorForToken("'!' unary expects bool type, found "+rightType.String(), tok)
		}
	default:
		p.appendErrorForToken("invalid unary operator", tok)
	}
}

func (p *parser) validateBinaryType(binaryExp *BinaryExpression) {
	tok := binaryExp.Token()
	op := binaryExp.Op
	if op == OP_ILLEGAL || op == OP_BANG {
		p.appendErrorForToken("invalid binary operator", tok)
		return
	}

	leftType := binaryExp.Left.Type()
	rightType := binaryExp.Right.Type()
	if !leftType.Matches(rightType) {
		p.appendErrorForToken("mismatched type for "+op.String()+": "+leftType.String()+", "+rightType.String(), tok)
		return
	}

	switch op {
	case OP_PLUS:
		if leftType != NUM_TYPE && leftType != STRING_TYPE && leftType.Name != ARRAY {
			p.appendErrorForToken("'+' takes num, string or array type, found "+leftType.String(), tok)
		}
	case OP_MINUS, OP_SLASH, OP_ASTERISK, OP_PERCENT:
		if leftType != NUM_TYPE {
			p.appendErrorForToken("'"+op.String()+"' takes num type, found "+leftType.String(), tok)
		}
	case OP_LT, OP_GT, OP_LTEQ, OP_GTEQ:
		if leftType != NUM_TYPE && leftType != STRING_TYPE {
			p.appendErrorForToken("'"+op.String()+"' takes num or string type, found "+leftType.String(), tok)
		}
	case OP_AND, OP_OR:
		if leftType != BOOL_TYPE {
			p.appendErrorForToken("'"+op.String()+"' takes bool type, found "+leftType.String(), tok)
		}
	}
}

func (p *parser) parseLiteral(scope *scope) Node {
	tok := p.cur
	tt := tok.TokenType()
	switch tt {
	case lexer.STRING_LIT:
		p.advance()
		return &StringLiteral{token: tok, Value: tok.Literal}
	case lexer.NUM_LIT:
		p.advance()
		val, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			p.appendError(err.Error())
			return nil
		}
		return &NumLiteral{token: tok, Value: val}
	case lexer.TRUE, lexer.FALSE:
		p.advance()
		return &Bool{token: tok, Value: tt == lexer.TRUE}
	case lexer.LBRACKET:
		return p.parseArrayLiteral(scope)
	case lexer.LCURLY:
		return p.parseMapLiteral(scope)
	}
	return nil
}

func (p *parser) parseArrayLiteral(scope *scope) Node {
	tok := p.cur
	p.advance()                   // advance past [
	multi := p.parseMulitlineWS() // allow whitespace after `[`, eg [ 1 2 3 ]
	elements := []Node{}
	tt := p.cur.TokenType()
	for tt != lexer.RBRACKET && tt != lexer.EOF {
		n := p.parseExprWSS(scope)
		if n == nil {
			return nil // previous error
		}
		elements = append(elements, n)
		multi = append(multi, multilineEl)
		multi = append(multi, p.parseMulitlineWS()...)
		tt = p.cur.TokenType()
	}
	if !p.assertToken(lexer.RBRACKET) {
		return nil
	}
	p.advance() // advance past ]
	arrayLit := &ArrayLiteral{token: tok, T: GENERIC_ARRAY}
	p.formatting.recordMultiline(arrayLit, multi)
	if len(elements) == 0 {
		return arrayLit
	}
	types := make([]*Type, len(elements))
	for i, e := range elements {
		types[i] = e.Type()
	}
	arrayLit.T = &Type{Name: ARRAY, Sub: p.combineTypes(types)}
	arrayLit.Elements = elements
	return arrayLit
}

func (p *parser) parseExprList(scope *scope) []Node {
	list := []Node{}
	tt := p.cur.TokenType()
	for tt != lexer.RPAREN && tt != lexer.RBRACKET && tt != lexer.EOF && !p.isAtEOL() {
		n := p.parseExprWSS(scope)
		if n == nil {
			return nil // previous error
		}
		list = append(list, n)
		p.advanceIfWS()
		tt = p.cur.TokenType()
	}
	return list
}

func (p *parser) parseExprWSS(scope *scope) Node {
	p.pushWSS(true)
	defer p.popWSS()
	return p.parseExpr(scope, LOWEST)
}

func (p *parser) combineTypes(types []*Type) *Type {
	combinedT := types[0]
	for _, t := range types[1:] {
		if combinedT.Accepts(t) {
			continue
		}
		if t.Accepts(combinedT) {
			combinedT = t
			continue
		}
		return ANY_TYPE
	}
	return combinedT
}

func (p *parser) parseMapLiteral(scope *scope) Node {
	p.pushWSS(false)
	defer p.popWSS()
	tok := p.cur
	mapLit := &MapLiteral{token: tok, Pairs: map[string]Node{}, T: GENERIC_MAP}
	p.advance() // advance past {

	if ok := p.parseMapPairs(scope, mapLit); !ok {
		return nil // previous error
	}
	if !p.assertToken(lexer.RCURLY) {
		return nil
	}
	p.advance() // advance past }
	if len(mapLit.Pairs) == 0 {
		return mapLit
	}
	types := make([]*Type, 0, len(mapLit.Pairs))
	for _, n := range mapLit.Pairs {
		types = append(types, n.Type())
	}
	mapLit.T = &Type{Name: MAP, Sub: p.combineTypes(types)}
	return mapLit
}

func (p *parser) parseMapPairs(scope *scope, mapLit *MapLiteral) bool {
	multi := p.parseMulitlineWS()
	tt := p.cur.TokenType()

	for tt != lexer.RCURLY && tt != lexer.EOF {
		if tt != lexer.IDENT {
			p.appendError("expected map key, found " + p.cur.FormatDetails())
		}
		key := p.cur.Literal
		p.advance() // advance past key IDENT
		if _, ok := mapLit.Pairs[key]; ok {
			p.appendError("duplicated map key'" + key + "'")
			return false
		}
		p.assertToken(lexer.COLON)
		p.advance() // advance past COLON

		n := p.parseExprWSS(scope)
		if n == nil {
			return false // previous error
		}
		mapLit.Pairs[key] = n
		mapLit.Order = append(mapLit.Order, key)
		multi = append(multi, multilineItem(key))
		multi = append(multi, p.parseMulitlineWS()...)
		tt = p.cur.TokenType()
	}
	p.formatting.recordMultiline(mapLit, multi)
	return true
}

// lookupVar looks up current token literal (IDENT) in scope.
// it assumes use, meaning reading of the variable, by marking the
// variable as used and hinting at using () around function calls.
// Do not use for writes, e.g. in left side of assignment.
func (p *parser) lookupVar(scope *scope) Node {
	tok := p.cur
	name := p.cur.Literal
	p.advance()
	if name == "_" {
		p.appendErrorForToken("anonymous variable '_' cannot be read", tok)
		return nil
	}
	if v, ok := scope.get(name); ok {
		v.isUsed = true
		return v
	}
	if _, ok := p.funcs[name]; ok {
		p.appendErrorForToken("function call must be parenthesized: ("+name+" ...)", tok)
		return nil
	}
	p.appendErrorForToken("unknown variable name '"+name+"'", tok)
	return nil
}
