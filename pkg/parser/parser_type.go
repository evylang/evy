package parser

import "foxygo.at/evy/pkg/lexer"

// parseType parses `num[]{}` into `MAP ARRAY NUM` inverting the order.
func (p *Parser) parseType() *Type {
	result := p.parseBasicType()
	if result == ILLEGAL_TYPE {
		return result
	}
	return p.parseSubType(result)
}

func (p *Parser) parseBasicType() *Type {
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
	}
	return ILLEGAL_TYPE
}

func (p *Parser) parseSubType(parent *Type) *Type {
	tt := p.cur.TokenType()
	typeName := compositeTypeName(tt)
	if typeName == ILLEGAL { // we have moved passed the type declaration
		return parent
	}
	if !matchParen(tt, p.peek.Type) {
		return ILLEGAL_TYPE
	}
	p.advance() // advance past opening token `[` or `{`
	p.advance() // advance past closing token `]` or `}`
	node := &Type{Name: typeName, Sub: parent}
	return p.parseSubType(node)
}
