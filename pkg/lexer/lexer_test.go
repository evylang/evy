package lexer

import (
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestSingleToken(t *testing.T) {
	tests := []struct {
		in   string
		want TokenType
	}{
		{in: ":=", want: DECLARE}, {in: "=", want: ASSIGN}, {in: "+", want: PLUS}, {in: "-", want: MINUS}, {in: "!", want: BANG}, {in: "*", want: ASTERISK}, {in: "/", want: SLASH},
		{in: "==", want: EQ}, {in: "!=", want: NOT_EQ}, {in: "<", want: LT}, {in: ">", want: GT}, {in: "<=", want: LTEQ}, {in: ">=", want: GTEQ},
		{in: "(", want: LPAREN}, {in: ")", want: RPAREN}, {in: "[", want: LBRACKET}, {in: "]", want: RBRACKET}, {in: "{", want: LCURLY}, {in: "}", want: RCURLY},
		{in: ":", want: COLON}, {in: ".", want: DOT}, {in: "...", want: DOT3},
		{in: "num", want: NUM}, {in: "string", want: STRING}, {in: "bool", want: BOOL}, {in: "any", want: ANY},
		{in: "true", want: TRUE}, {in: "false", want: FALSE}, {in: "and", want: AND}, {in: "or", want: OR},
		{in: "if", want: IF}, {in: "else", want: ELSE}, {in: "func", want: FUNC}, {in: "return", want: RETURN}, {in: "for", want: FOR}, {in: "range", want: RANGE}, {in: "while", want: WHILE}, {in: "break", want: BREAK}, {in: "end", want: END},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			l := New(tt.in)
			got := l.Next()
			want := &Token{Type: tt.want, Literal: "", Offset: 0, Line: 1, Col: 1}
			assert.Equal(t, want, got)

			wantFormat := tt.in
			gotFormat := got.Format()
			assert.Equal(t, wantFormat, gotFormat)

			got = l.Next()
			offset := len(tt.in)
			want = &Token{Type: EOF, Literal: "", Offset: offset, Line: 1, Col: offset + 1}
			assert.Equal(t, want, got)
		})
	}
}

func TestSingleTokenWithLiteral(t *testing.T) {
	tests := map[string]struct {
		in   string
		want *Token
	}{
		"IDENT": {
			in:   "x",
			want: &Token{Type: IDENT, Literal: "x", Offset: 0, Line: 1, Col: 1},
		}, "IDENT whitespace": {
			in:   "  x2 ",
			want: &Token{Type: IDENT, Literal: "x2", Offset: 2, Line: 1, Col: 3},
		}, "STRING": {
			in:   `"xy"`,
			want: &Token{Type: STRING_LIT, Literal: "xy", Offset: 0, Line: 1, Col: 1},
		}, "STRING whitespace": {
			in:   ` " x y "  `,
			want: &Token{Type: STRING_LIT, Literal: " x y ", Offset: 1, Line: 1, Col: 2},
		}, "STRING WITH COMMENT TOKEN": {
			in:   `"x//y"`,
			want: &Token{Type: STRING_LIT, Literal: "x//y", Offset: 0, Line: 1, Col: 1},
		}, "STRING escaped quote": {
			in:   `"xy\""`,
			want: &Token{Type: STRING_LIT, Literal: `xy"`, Offset: 0, Line: 1, Col: 1},
		}, "NUM": {
			in:   " 1  ",
			want: &Token{Type: NUM_LIT, Literal: "1", Offset: 1, Line: 1, Col: 2},
		}, "NUM2": {
			in:   "12 ",
			want: &Token{Type: NUM_LIT, Literal: "12", Offset: 0, Line: 1, Col: 1},
		}, "NUM3": {
			in:   "12.3 ",
			want: &Token{Type: NUM_LIT, Literal: "12.3", Offset: 0, Line: 1, Col: 1},
		}, "NUM4": {
			in:   "12.3.4 ",
			want: &Token{Type: NUM_LIT, Literal: "12.3.4", Offset: 0, Line: 1, Col: 1},
		}, "COMMENT": {
			in:   "  // x2 ",
			want: &Token{Type: COMMENT, Literal: "// x2 ", Offset: 2, Line: 1, Col: 3},
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			l := New(tt.in)
			got := l.Next()
			assert.Equal(t, tt.want, got)

			eofOffset := len(tt.in)
			want := &Token{Type: EOF, Literal: "", Offset: eofOffset, Line: 1, Col: eofOffset + 1}
			got = l.Next()
			assert.Equal(t, want, got)
		})
	}
}

func TestNums(t *testing.T) {
	in := `
y := 1
y = y * ( 3 + 7)
`
	wantTokens := []Token{
		{Type: NL, Literal: "", Offset: 0, Line: 1, Col: 1},
		{Type: IDENT, Literal: "y", Offset: 1, Line: 2, Col: 1},
		{Type: DECLARE, Literal: "", Offset: 3, Line: 2, Col: 3},
		{Type: NUM_LIT, Literal: "1", Offset: 6, Line: 2, Col: 6},
		{Type: NL, Literal: "", Offset: 7, Line: 2, Col: 7},
		{Type: IDENT, Literal: "y", Offset: 8, Line: 3, Col: 1},
		{Type: ASSIGN, Literal: "", Offset: 10, Line: 3, Col: 3},
		{Type: IDENT, Literal: "y", Offset: 12, Line: 3, Col: 5},
		{Type: ASTERISK, Literal: "", Offset: 14, Line: 3, Col: 7},
		{Type: LPAREN, Literal: "", Offset: 16, Line: 3, Col: 9},
		{Type: NUM_LIT, Literal: "3", Offset: 18, Line: 3, Col: 11},
		{Type: PLUS, Literal: "", Offset: 20, Line: 3, Col: 13},
		{Type: NUM_LIT, Literal: "7", Offset: 22, Line: 3, Col: 15},
		{Type: RPAREN, Literal: "", Offset: 23, Line: 3, Col: 16},
		{Type: NL, Literal: "", Offset: 24, Line: 3, Col: 17},
		{Type: EOF, Literal: "", Offset: len(in), Line: 4, Col: 1},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, &want, got)
	}
}

func TestStrings(t *testing.T) {
	in := `
s:string
s = "abc"
s = s[:1] // "bc"`
	wantTokens := []Token{
		{Type: NL, Literal: "", Offset: 0, Line: 1, Col: 1},
		{Type: IDENT, Literal: "s", Offset: 1, Line: 2, Col: 1},
		{Type: COLON, Literal: "", Offset: 2, Line: 2, Col: 2},
		{Type: STRING, Literal: "", Offset: 3, Line: 2, Col: 3},
		{Type: NL, Literal: "", Offset: 9, Line: 2, Col: 9},

		{Type: IDENT, Literal: "s", Offset: 10, Line: 3, Col: 1},
		{Type: ASSIGN, Literal: "", Offset: 12, Line: 3, Col: 3},
		{Type: STRING_LIT, Literal: "abc", Offset: 14, Line: 3, Col: 5},
		{Type: NL, Literal: "", Offset: 19, Line: 3, Col: 10},

		{Type: IDENT, Literal: "s", Offset: 20, Line: 4, Col: 1},
		{Type: ASSIGN, Literal: "", Offset: 22, Line: 4, Col: 3},
		{Type: IDENT, Literal: "s", Offset: 24, Line: 4, Col: 5},
		{Type: LBRACKET, Literal: "", Offset: 25, Line: 4, Col: 6},
		{Type: COLON, Literal: "", Offset: 26, Line: 4, Col: 7},
		{Type: NUM_LIT, Literal: "1", Offset: 27, Line: 4, Col: 8},
		{Type: RBRACKET, Literal: "", Offset: 28, Line: 4, Col: 9},
		{Type: COMMENT, Literal: `// "bc"`, Offset: 30, Line: 4, Col: 11},
		{Type: EOF, Literal: "", Offset: len(in), Line: 4, Col: 18},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, &want, got)
	}
}

func TestMapRange(t *testing.T) {
	in := `
m := string{ name:"Mali" sport:"climbing" }
for key := range m
    print key m[key] // Mali climbing
end
`
	wantTokens := []Token{
		{Type: NL, Literal: "", Offset: 0, Line: 1, Col: 1},
		{Type: IDENT, Literal: "m", Offset: 1, Line: 2, Col: 1},
		{Type: DECLARE, Literal: "", Offset: 3, Line: 2, Col: 3},
		{Type: STRING, Literal: "", Offset: 6, Line: 2, Col: 6},
		{Type: LCURLY, Literal: "", Offset: 12, Line: 2, Col: 12},
		{Type: IDENT, Literal: "name", Offset: 14, Line: 2, Col: 14},
		{Type: COLON, Literal: "", Offset: 18, Line: 2, Col: 18},
		{Type: STRING_LIT, Literal: "Mali", Offset: 19, Line: 2, Col: 19},
		{Type: IDENT, Literal: "sport", Offset: 26, Line: 2, Col: 26},
		{Type: COLON, Literal: "", Offset: 31, Line: 2, Col: 31},
		{Type: STRING_LIT, Literal: "climbing", Offset: 32, Line: 2, Col: 32},
		{Type: RCURLY, Literal: "", Offset: 43, Line: 2, Col: 43},
		{Type: NL, Literal: "", Offset: 44, Line: 2, Col: 44},

		{Type: FOR, Literal: "", Offset: 45, Line: 3, Col: 1},
		{Type: IDENT, Literal: "key", Offset: 49, Line: 3, Col: 5},
		{Type: DECLARE, Literal: "", Offset: 53, Line: 3, Col: 9},
		{Type: RANGE, Literal: "", Offset: 56, Line: 3, Col: 12},
		{Type: IDENT, Literal: "m", Offset: 62, Line: 3, Col: 18},
		{Type: NL, Literal: "", Offset: 63, Line: 3, Col: 19},

		{Type: IDENT, Literal: "print", Offset: 68, Line: 4, Col: 5},
		{Type: IDENT, Literal: "key", Offset: 74, Line: 4, Col: 11},
		{Type: IDENT, Literal: "m", Offset: 78, Line: 4, Col: 15},
		{Type: LBRACKET, Literal: "", Offset: 79, Line: 4, Col: 16},
		{Type: IDENT, Literal: "key", Offset: 80, Line: 4, Col: 17},
		{Type: RBRACKET, Literal: "", Offset: 83, Line: 4, Col: 20},
		{Type: COMMENT, Literal: "// Mali climbing", Offset: 85, Line: 4, Col: 22},
		{Type: NL, Literal: "", Offset: 101, Line: 4, Col: 38},

		{Type: END, Literal: "", Offset: 102, Line: 5, Col: 1},
		{Type: NL, Literal: "", Offset: 105, Line: 5, Col: 4},

		{Type: EOF, Literal: "", Offset: len(in), Line: 6, Col: 1},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, &want, got)
	}
}

func TestIllegal(t *testing.T) {
	tests := []struct {
		in   string
		want Token
	}{
		{in: "  , ", want: Token{Offset: 2, Literal: ",", Line: 1, Col: 3}},
		{in: `"unterminated`, want: Token{Offset: 0, Literal: `"`, Line: 1, Col: 1}},
		{in: ` "newline in the
		middle "`, want: Token{Offset: 1, Literal: `"`, Line: 1, Col: 2}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			l := New(tt.in)
			got := l.Next()
			assert.Equal(t, &tt.want, got)
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: ":=", want: "DECLARE\n"},
		{in: "", want: ""},
		{in: "  ", want: ""},
		{in: "\n", want: "NL\n"},
		{in: " \n ", want: "NL\n"},
		{in: ` "abc"`, want: "STRING_LIT 'abc'\n"},
		{in: ",", want: "ILLEGAL 💥 ',' at line 1 column 1\n"},
		{in: `"asdf `, want: "ILLEGAL 💥 '\"' at line 1 column 1\n"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			got := Run(tt.in)
			want := tt.want + "EOF\n"
			assert.Equal(t, want, got)
		})
	}
}
