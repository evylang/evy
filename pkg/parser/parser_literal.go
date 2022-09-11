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

func (p *Parser) parseLiteral() Node {
	tok := p.cur
	tt := tok.TokenType()
	p.advance()
	switch tt {
	case lexer.STRING_LIT:
		return &StringLiteral{Token: tok, Value: tok.Literal}
	case lexer.NUM_LIT:
		val, err := strconv.ParseFloat(tok.Literal, 64)
		if err != nil {
			p.appendError(err.Error())
			return nil
		}
		return &NumLiteral{Token: tok, Value: val}
	case lexer.TRUE, lexer.FALSE:
		return &Bool{Token: tok, Value: tt == lexer.TRUE}
	}
	return nil
}
