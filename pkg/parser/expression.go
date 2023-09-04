// This file contains a Top Down Operator Precedence or Pratt parser.
//
// It is based on Thorston Ball's monkey interpreter:
// https://github.com/juliaogris/monkey/blob/master/parser/parser.go
//
// The expression parser is implemented in receiver functions of the
// Parser struct as defined in parser.go.

package parser

import (
	"fmt"
	"strconv"

	"foxygo.at/evy/pkg/lexer"
)

type precedence int

const (
	lowestPrec      precedence = iota
	orPrec                     // or
	andPrec                    // and
	equalsPrec                 // ==
	lessgreaterPrec            // > or <
	sumPrec                    // +
	productPrec                // *
	unaryPrec                  // -x  !x
	indexPrec                  // array[i]
)

var precedences = map[lexer.TokenType]precedence{
	lexer.EQ:       equalsPrec,
	lexer.NOT_EQ:   equalsPrec,
	lexer.LT:       lessgreaterPrec,
	lexer.GT:       lessgreaterPrec,
	lexer.LTEQ:     lessgreaterPrec,
	lexer.GTEQ:     lessgreaterPrec,
	lexer.PLUS:     sumPrec,
	lexer.MINUS:    sumPrec,
	lexer.OR:       orPrec,
	lexer.SLASH:    productPrec,
	lexer.ASTERISK: productPrec,
	lexer.PERCENT:  productPrec,
	lexer.AND:      andPrec,
	lexer.LBRACKET: indexPrec,
	lexer.DOT:      indexPrec,
}

func (p *parser) parseTopLevelExpr() Node {
	tok := p.cur
	if tok.Type == lexer.IDENT && p.funcs[tok.Literal] != nil {
		return p.parseFuncCall()
	}
	return p.parseExpr(lowestPrec)
}

func (p *parser) parseFuncCall() Node {
	fc := &FuncCall{token: p.cur, Name: p.cur.Literal}
	p.advance() // advance past function name IDENT
	funcDef := p.funcs[fc.Name]
	funcDef.isCalled = true
	fc.FuncDef = funcDef
	fc.Arguments = p.parseExprList()
	p.assertArgTypes(fc.FuncDef, fc.Arguments)
	return fc
}

func (p *parser) parseExpr(prec precedence) Node {
	var left Node
	switch p.cur.Type {
	case lexer.IDENT:
		left = p.lookupVar()
	case lexer.STRING_LIT, lexer.NUM_LIT, lexer.TRUE, lexer.FALSE, lexer.LBRACKET, lexer.LCURLY:
		left = p.parseLiteral()
	case lexer.BANG, lexer.MINUS:
		left = p.parseUnaryExpr()
	case lexer.LPAREN:
		left = p.parseGroupedExpr()
	default:
		p.unexpectedLeftTokenError()
	}
	for left != nil && !p.isAtExprEnd() && prec < precedences[p.cur.Type] {
		tt := p.cur.Type
		switch {
		case isBinaryOp(tt):
			left = p.parseBinaryExpr(left)
		case tt == lexer.LBRACKET:
			left = p.parseIndexOrSliceExpr(left, true)
		case tt == lexer.DOT && p.peek.Type == lexer.LPAREN:
			left = p.parseTypeAssertion(left)
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
			p.appendError("unexpected whitespace before " + p.cur.Format())
			return
		}
		if tt == lexer.WS && isBinaryOp(prevTT) {
			prevToken := p.lookAt(p.pos - 1)
			p.appendErrorForToken("unexpected whitespace after "+prevToken.Format(), prevToken)
			return
		}
	}
	p.appendError("unexpected " + p.cur.Format())
}

func (p *parser) isAtExprEnd() bool {
	if p.isWSS() && p.cur.Type == lexer.WS {
		return true
	}
	return p.isAtEOL()
}

func (p *parser) parseUnaryExpr() Node {
	tok := p.cur
	unaryExp := &UnaryExpression{token: tok, Op: op(tok)}
	p.advance() // advance past operator
	if p.lookAt(p.pos-1).Type == lexer.WS {
		msg := fmt.Sprintf("unexpected whitespace after %q", unaryExp.Op.String())
		p.appendErrorForToken(msg, tok)
	}
	unaryExp.Right = p.parseExpr(unaryPrec)
	if unaryExp.Right == nil {
		return nil // previous error
	}
	p.validateUnaryType(unaryExp)
	return unaryExp
}

func (p *parser) parseBinaryExpr(left Node) Node {
	tok := p.cur
	expType := left.Type()
	if isComparisonOp(tok.Type) {
		expType = BOOL_TYPE
	}
	binaryExp := &BinaryExpression{token: tok, T: expType, Op: op(tok), Left: left}
	prec := precedences[tok.Type]
	p.advance() // advance past operator
	binaryExp.Right = p.parseExpr(prec)
	if binaryExp.Right == nil {
		return nil // previous error
	}
	if expType == UNTYPED_ARRAY {
		binaryExp.T = binaryExp.Right.Type() // array concatenation e.g. [] + [1 2]
	}
	p.validateBinaryType(binaryExp)
	if p.isWSS() {
		p.formatting.recordWSS(binaryExp)
	}

	return binaryExp
}

func (p *parser) parseGroupedExpr() Node {
	p.pushWSS(false)
	defer p.popWSS()
	tok := p.cur
	p.advance() // advance past (
	exp := p.parseTopLevelExpr()
	if !p.assertToken(lexer.RPAREN) || exp == nil {
		return nil
	}
	p.advanceWSS() // advance past )
	return &GroupExpression{token: tok, Expr: exp}
}

func (p *parser) parseIndexOrSliceExpr(left Node, allowSlice bool) Node {
	p.pushWSS(false)
	defer p.popWSS()
	tok := p.cur
	if p.lookAt(p.pos-1).Type == lexer.WS {
		p.appendError(`unexpected whitespace before "["`)
		return nil
	}
	p.advance() // advance past [
	leftType := left.Type().Name
	if leftType != ARRAY && leftType != MAP && leftType != STRING {
		p.appendErrorForToken("only array, string and map type can be indexed, found "+left.Type().String(), tok)
		return nil
	}
	if p.cur.TokenType() == lexer.COLON && allowSlice { // e.g. a[:2]
		p.advance() //  advance past :
		return p.parseSlice(tok, left, nil)
	}
	index := p.parseTopLevelExpr()
	if index == nil {
		return nil
	}
	tt := p.cur.TokenType()
	if tt == lexer.COLON && allowSlice { // e.g. a[1:3] or a[1:]
		p.advance() // advance past :
		return p.parseSlice(tok, left, index)
	}
	if !p.validateIndex(tok, leftType, index.Type()) {
		return nil
	}
	p.advanceWSS() // advance past ]
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
		p.appendErrorForToken(leftType.name()+" index expects num, found "+indexType.String(), tok)
		return false
	}
	if leftType == MAP && indexType != STRING_TYPE {
		p.appendErrorForToken("map index expects string, found "+indexType.String(), tok)
		return false
	}
	return true
}

func (p *parser) parseSlice(tok *lexer.Token, left, start Node) Node {
	leftType := left.Type().Name
	if leftType != ARRAY && leftType != STRING {
		p.appendErrorForToken("only array and string can be sliced, found "+left.Type().String(), tok)
		return nil
	}

	t := left.Type()
	if p.cur.Type == lexer.RBRACKET {
		p.advance()
		return &SliceExpression{token: tok, Left: left, Start: start, End: nil, T: t}
	}
	end := p.parseTopLevelExpr()
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
		p.appendError(`unexpected whitespace before "."`)
		return nil
	}
	if p.lookAt(p.pos+1).Type == lexer.WS {
		p.appendError(`unexpected whitespace after "."`)
		return nil
	}
	p.advance() // advance past .
	leftType := left.Type().Name
	if leftType != MAP {
		p.appendErrorForToken(`field access with "." expects map type, found `+left.Type().String(), tok)
		return nil
	}
	if p.cur.TokenType() != lexer.IDENT {
		p.appendErrorForToken(`expected map key, found `+p.cur.TokenType().String(), tok)
		return nil
	}
	expr := &DotExpression{token: tok, Left: left, T: left.Type().Sub, Key: p.cur.Literal}
	p.advance() // advance past key IDENT
	return expr
}

func (p *parser) parseTypeAssertion(left Node) Node {
	tok := p.cur
	if p.lookAt(p.pos-1).Type == lexer.WS {
		p.appendError(`unexpected whitespace before "."`)
		return nil
	}
	if p.lookAt(p.pos+1).Type == lexer.WS {
		p.appendError(`unexpected whitespace after "."`)
		return nil
	}
	p.pushWSS(false)
	defer p.popWSS()
	p.advance() // advance past .
	p.advance() // advance past (
	t := p.parseType()
	switch t {
	case ILLEGAL_TYPE:
		msg := fmt.Sprintf("invalid type in type assertion of %q", left.String())
		p.appendErrorForToken(msg, tok)
	case ANY_TYPE:
		p.appendErrorForToken("cannot type assert to type any", tok)
	}
	if p.assertToken(lexer.RPAREN) {
		p.advanceWSS() // advance past )
	}
	if left.Type() != ANY_TYPE {
		p.appendErrorForToken("value of type assertion must be of type any, not "+left.Type().String(), tok)
	}
	return &TypeAssertion{T: t, token: tok, Left: left}
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
			p.appendErrorForToken(`"-" unary expects num type, found `+rightType.String(), tok)
		}
	case OP_BANG:
		if unaryExp.Right.Type() != BOOL_TYPE {
			p.appendErrorForToken(`"!" unary expects bool type, found `+rightType.String(), tok)
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
	if !leftType.matches(rightType) {
		msg := fmt.Sprintf("mismatched type for %s: %s, %s", op.String(), leftType.String(), rightType.String())
		p.appendErrorForToken(msg, tok)
		return
	}

	switch op {
	case OP_PLUS:
		if leftType != NUM_TYPE && leftType != STRING_TYPE && leftType.Name != ARRAY {
			p.appendErrorForToken(`"+" takes num, string or array type, found `+leftType.String(), tok)
		}
	case OP_MINUS, OP_SLASH, OP_ASTERISK, OP_PERCENT:
		if leftType != NUM_TYPE {
			msg := fmt.Sprintf("%q takes num type, found %s", op.String(), leftType.String())
			p.appendErrorForToken(msg, tok)
		}
	case OP_LT, OP_GT, OP_LTEQ, OP_GTEQ:
		if leftType != NUM_TYPE && leftType != STRING_TYPE {
			msg := fmt.Sprintf("%q takes num or string type, found %s", op.String(), leftType.String())
			p.appendErrorForToken(msg, tok)
		}
	case OP_AND, OP_OR:
		if leftType != BOOL_TYPE {
			msg := fmt.Sprintf("%q takes bool type, found %s", op.String(), leftType.String())
			p.appendErrorForToken(msg, tok)
		}
	}
}

func (p *parser) parseLiteral() Node {
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
		return &BoolLiteral{token: tok, Value: tt == lexer.TRUE}
	case lexer.LBRACKET:
		return p.parseArrayLiteral()
	case lexer.LCURLY:
		return p.parseMapLiteral()
	}
	return nil
}

func (p *parser) parseArrayLiteral() Node {
	tok := p.cur
	p.advance()                   // advance past [
	multi := p.parseMulitlineWS() // allow whitespace after `[`, eg [ 1 2 3 ]
	elements := []Node{}
	tt := p.cur.TokenType()
	for tt != lexer.RBRACKET && tt != lexer.EOF {
		n := p.parseExprWSS()
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
	arrayLit := &ArrayLiteral{token: tok, T: UNTYPED_ARRAY}
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

func (p *parser) parseExprList() []Node {
	list := []Node{}
	tt := p.cur.TokenType()
	for tt != lexer.RPAREN && tt != lexer.RBRACKET && tt != lexer.EOF && !p.isAtEOL() {
		n := p.parseExprWSS()
		if n == nil {
			return nil // previous error
		}
		list = append(list, n)
		p.advanceIfWS()
		tt = p.cur.TokenType()
	}
	return list
}

func (p *parser) parseExprWSS() Node {
	p.pushWSS(true)
	defer p.popWSS()
	return p.parseExpr(lowestPrec)
}

func (p *parser) combineTypes(types []*Type) *Type {
	combinedT := types[0]
	for _, t := range types[1:] {
		if combinedT.accepts(t) {
			if combinedT.IsUntyped() {
				combinedT = t
			}
			continue
		}
		// same composite types can be combined, for instance
		// []string and []num become []any in
		// {a:["X" "Y"] b:[1 2]}
		if t.sameComposite(combinedT) {
			combinedT = &Type{Name: t.Name, Sub: ANY_TYPE}
			continue
		}
		return ANY_TYPE
	}
	return combinedT
}

func (p *parser) parseMapLiteral() Node {
	p.pushWSS(false)
	defer p.popWSS()
	tok := p.cur
	mapLit := &MapLiteral{token: tok, Pairs: map[string]Node{}, T: UNTYPED_MAP}
	p.advance() // advance past {

	if ok := p.parseMapPairs(mapLit); !ok {
		return nil // previous error
	}
	if !p.assertToken(lexer.RCURLY) {
		return nil
	}
	p.advanceWSS() // advance past }
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

func (p *parser) parseMapPairs(mapLit *MapLiteral) bool {
	multi := p.parseMulitlineWS()
	tt := p.cur.TokenType()

	for tt != lexer.RCURLY && tt != lexer.EOF {
		if tt != lexer.IDENT {
			p.appendError("expected map key, found " + p.cur.Format())
		}
		key := p.cur.Literal
		p.advance() // advance past key IDENT
		if _, ok := mapLit.Pairs[key]; ok {
			p.appendError(fmt.Sprintf("duplicated map key %q", key))
			return false
		}
		p.assertToken(lexer.COLON)
		p.advance() // advance past COLON

		n := p.parseExprWSS()
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
func (p *parser) lookupVar() Node {
	tok := p.cur
	name := p.cur.Literal
	p.advance()
	if name == "_" {
		p.appendErrorForToken(`anonymous variable "_" cannot be read`, tok)
		return nil
	}
	if v, ok := p.scope.get(name); ok {
		v.isUsed = true
		v2 := *v
		v2.token = tok
		return &v2
	}
	if _, ok := p.funcs[name]; ok {
		msg := fmt.Sprintf("function call must be parenthesized: (%s ...)", name)
		p.appendErrorForToken(msg, tok)
		return nil
	}
	msg := fmt.Sprintf("unknown variable name %q", name)
	p.appendErrorForToken(msg, tok)
	return nil
}
