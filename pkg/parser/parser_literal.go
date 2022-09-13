package parser

import (
	"strconv"

	"foxygo.at/evy/pkg/lexer"
)

func (p *Parser) isLiteral() bool {
	tt := p.cur.TokenType()
	if tt == lexer.NUM_LIT || tt == lexer.STRING_LIT || tt == lexer.TRUE || tt == lexer.FALSE {
		return true
	}
	if !isBasicType(tt) {
		return false
	}
	peek := p.peek.TokenType()
	return peek == lexer.LBRACKET || peek == lexer.LCURLY
}

func (p *Parser) parseLiteral(scope *scope) Node {
	tok := p.cur
	tt := tok.TokenType()
	switch tt {
	case lexer.STRING_LIT:
		p.advance()
		return &StringLiteral{Token: tok, Value: tok.Literal}
	case lexer.NUM_LIT:
		p.advance()
		val, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			p.appendError(err.Error())
			return nil
		}
		return &NumLiteral{Token: tok, Value: val}
	case lexer.TRUE, lexer.FALSE:
		p.advance()
		return &Bool{Token: tok, Value: tt == lexer.TRUE}
	}
	return p.parseCompositeLiteral(scope)
}

func (p *Parser) parseCompositeLiteral(scope *scope) Node {
	tok := p.cur
	litType := p.parseLiteralType()
	switch litType.Name {
	case ARRAY:
		elements := p.parseArrayElements(scope, litType.Sub)
		return &ArrayLiteral{Token: tok, Elements: elements, nType: litType}
	case MAP:
		pairs, order := p.parseMapPairs(scope, litType.Sub)
		return &MapLiteral{
			Token: tok,
			Pairs: pairs,
			Order: order,
			nType: litType,
		}
	}
	p.appendError("unknown literal " + tok.String())
	return nil
}

func (p *Parser) parseArrayElements(scope *scope, t *Type) []Node {
	terms := p.parseTerms(scope)
	tt := p.cur.TokenType()
	p.advance()
	if tt != lexer.RBRACKET {
		p.appendError("unterminated array literal")
		return nil
	}
	for _, term := range terms {
		if !t.Accepts(term.Type()) {
			p.appendError("array literal '" + term.String() + "' should have type '" + t.Format() + "'")
		}
	}
	return terms
}

func (p *Parser) parseMapPairs(scope *scope, t *Type) (map[string]Node, []string) {
	pairs := map[string]Node{}
	var order []string
	for !p.isTermsEnd() {
		tt := p.cur.TokenType()
		key := p.cur.Literal
		p.advance()
		p.assertToken(lexer.COLON)
		p.advance()
		val := p.parseTerm(scope)
		if tt != lexer.IDENT {
			p.appendError("invalid map key '" + tt.Format() + "'")
			continue
		}
		if _, ok := pairs[key]; ok {
			p.appendError("duplicated map key'" + key + "'")
			continue
		}
		// type check
		if !t.Accepts(val.Type()) {
			p.appendError("map literal '" + val.String() + "' should have type '" + t.Format() + "'")
		}

		pairs[key] = val
		order = append(order, key)
	}
	tt := p.cur.TokenType()
	p.advance()
	if tt != lexer.RCURLY {
		p.appendError("unterminated map literal")
		return nil, nil
	}
	return pairs, order
}
