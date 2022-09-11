package parser

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
	t := basicTypeName(tt)
	p.advance()
	if t == ILLEGAL {
		return ILLEGAL_TYPE
	}
	return &Type{Name: t}
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

// parseLiteralType parses `num{}[` into `ARRAY MAP NUM` inverting the order.
// Parsing stops after the final opening rune `{` or `[`.
func (p *Parser) parseLiteralType() *Type {
	result := p.parseBasicType()
	if result == ILLEGAL_TYPE {
		return result
	}
	return p.parseSubTypeNoClose(result)
}

func (p *Parser) parseSubTypeNoClose(parent *Type) *Type {
	tt := p.cur.TokenType()
	typeName := compositeTypeName(tt)
	if typeName == ILLEGAL {
		return ILLEGAL_TYPE
	}
	// non-empty declaration `num[1]`
	if !matchParen(tt, p.peek.TokenType()) { // we have moved past the declaration
		p.advance() // advance past opening token `[` or `{`
		return &Type{Name: typeName, Sub: parent}
	}
	// empty declaration `num[]`
	peek2 := p.lookAt(p.pos + 2).TokenType()
	if compositeTypeName(peek2) == ILLEGAL {
		p.advance() // advance past opening token `[` or `{`
		return &Type{Name: typeName, Sub: parent}
	}
	// nested declaration `num[]{}`
	p.advance() // advance past opening token `[` or `{`
	p.advance() // advance past closing token `]` or `}`
	node := &Type{Name: typeName, Sub: parent}
	return p.parseSubTypeNoClose(node)
}
