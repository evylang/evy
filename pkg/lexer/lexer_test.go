package lexer

import (
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestSingleToken(t *testing.T) {
	tests := []struct {
		in   string
		want TokenType
	}{
		{in: ":=", want: DECLARE},
		{in: "=", want: ASSIGN},
		{in: "+", want: PLUS},
		{in: "-", want: MINUS},
		{in: "!", want: BANG},
		{in: "*", want: ASTERISK},
		{in: "/", want: SLASH},
		{in: "%", want: PERCENT},
		{in: "==", want: EQ},
		{in: "!=", want: NOT_EQ},
		{in: "<", want: LT},
		{in: ">", want: GT},
		{in: "<=", want: LTEQ},
		{in: ">=", want: GTEQ},
		{in: "(", want: LPAREN},
		{in: ")", want: RPAREN},
		{in: "[", want: LBRACKET},
		{in: "]", want: RBRACKET},
		{in: "{", want: LCURLY},
		{in: "}", want: RCURLY},
		{in: ":", want: COLON},
		{in: ".", want: DOT},
		{in: "...", want: DOT3},
		{in: "num", want: NUM},
		{in: "string", want: STRING},
		{in: "bool", want: BOOL},
		{in: "any", want: ANY},
		{in: "true", want: TRUE},
		{in: "false", want: FALSE},
		{in: "and", want: AND},
		{in: "or", want: OR},
		{in: "if", want: IF},
		{in: "else", want: ELSE},
		{in: "func", want: FUNC},
		{in: "return", want: RETURN},
		{in: "for", want: FOR},
		{in: "range", want: RANGE},
		{in: "while", want: WHILE},
		{in: "break", want: BREAK},
		{in: "end", want: END},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			l := New(tt.in)
			got := l.Next()
			want := &Token{Type: tt.want, Literal: "", Offset: 0, Line: 1, Col: 1}
			assert.Equal(t, want, got)

			wantFormat := `"` + tt.in + `"`
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
			in:   "x ",
			want: &Token{Type: IDENT, Literal: "x", Offset: 0, Line: 1, Col: 1},
		}, "IDENT whitespace": {
			in:   "x2\t",
			want: &Token{Type: IDENT, Literal: "x2", Offset: 0, Line: 1, Col: 1},
		}, "STRING": {
			in:   `"xy" `,
			want: &Token{Type: STRING_LIT, Literal: "xy", Offset: 0, Line: 1, Col: 1},
		}, "STRING whitespace": {
			in:   `" x y "  `,
			want: &Token{Type: STRING_LIT, Literal: " x y ", Offset: 0, Line: 1, Col: 1},
		}, "STRING WITH COMMENT TOKEN": {
			in:   `"x//y" `,
			want: &Token{Type: STRING_LIT, Literal: "x//y", Offset: 0, Line: 1, Col: 1},
		}, "STRING escaped quote": {
			in:   `"xy\""    `,
			want: &Token{Type: STRING_LIT, Literal: `xy"`, Offset: 0, Line: 1, Col: 1},
		}, "NUM": {
			in:   "1  \t ",
			want: &Token{Type: NUM_LIT, Literal: "1", Offset: 0, Line: 1, Col: 1},
		}, "NUM2": {
			in:   "12 ",
			want: &Token{Type: NUM_LIT, Literal: "12", Offset: 0, Line: 1, Col: 1},
		}, "NUM3": {
			in:   "12.3 ",
			want: &Token{Type: NUM_LIT, Literal: "12.3", Offset: 0, Line: 1, Col: 1},
		}, "NUM4": {
			in:   "12.3.4\t",
			want: &Token{Type: NUM_LIT, Literal: "12.3.4", Offset: 0, Line: 1, Col: 1},
		}, "COMMENT": {
			in:   "// x2 ",
			want: &Token{Type: COMMENT, Literal: "// x2 ", Offset: 0, Line: 1, Col: 1},
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			l := New(tt.in)
			got := l.Next()
			assert.Equal(t, tt.want, got)

			if name != "COMMENT" {
				got = l.Next()
				assert.Equal(t, WS, got.Type, name)
			}

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
y = y*(3 + 7)
`
	wantTokens := []*Token{
		{Type: NL, Literal: "", Offset: 0, Line: 1, Col: 1},
		{Type: IDENT, Literal: "y", Offset: 1, Line: 2, Col: 1},
		{Type: WS, Literal: "", Offset: 2, Line: 2, Col: 2},
		{Type: DECLARE, Literal: "", Offset: 3, Line: 2, Col: 3},
		{Type: WS, Literal: "", Offset: 5, Line: 2, Col: 5},
		{Type: NUM_LIT, Literal: "1", Offset: 6, Line: 2, Col: 6},
		{Type: NL, Literal: "", Offset: 7, Line: 2, Col: 7},
		{Type: IDENT, Literal: "y", Offset: 8, Line: 3, Col: 1},
		{Type: WS, Literal: "", Offset: 9, Line: 3, Col: 2},
		{Type: ASSIGN, Literal: "", Offset: 10, Line: 3, Col: 3},
		{Type: WS, Literal: "", Offset: 11, Line: 3, Col: 4},
		{Type: IDENT, Literal: "y", Offset: 12, Line: 3, Col: 5},
		{Type: ASTERISK, Literal: "", Offset: 13, Line: 3, Col: 6},
		{Type: LPAREN, Literal: "", Offset: 14, Line: 3, Col: 7},
		{Type: NUM_LIT, Literal: "3", Offset: 15, Line: 3, Col: 8},
		{Type: WS, Literal: "", Offset: 16, Line: 3, Col: 9},
		{Type: PLUS, Literal: "", Offset: 17, Line: 3, Col: 10},
		{Type: WS, Literal: "", Offset: 18, Line: 3, Col: 11},
		{Type: NUM_LIT, Literal: "7", Offset: 19, Line: 3, Col: 12},
		{Type: RPAREN, Literal: "", Offset: 20, Line: 3, Col: 13},
		{Type: NL, Literal: "", Offset: 21, Line: 3, Col: 14},
		{Type: EOF, Literal: "", Offset: len(in), Line: 4, Col: 1},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, want, got)
	}
}

func TestStrings(t *testing.T) {
	in := `
s:string
s = "abc"
s = s[:1] // "bc"`
	wantTokens := []*Token{
		{Type: NL, Literal: "", Offset: 0, Line: 1, Col: 1},
		{Type: IDENT, Literal: "s", Offset: 1, Line: 2, Col: 1},
		{Type: COLON, Literal: "", Offset: 2, Line: 2, Col: 2},
		{Type: STRING, Literal: "", Offset: 3, Line: 2, Col: 3},
		{Type: NL, Literal: "", Offset: 9, Line: 2, Col: 9},

		{Type: IDENT, Literal: "s", Offset: 10, Line: 3, Col: 1},
		{Type: WS, Literal: "", Offset: 11, Line: 3, Col: 2},
		{Type: ASSIGN, Literal: "", Offset: 12, Line: 3, Col: 3},
		{Type: WS, Literal: "", Offset: 13, Line: 3, Col: 4},
		{Type: STRING_LIT, Literal: "abc", Offset: 14, Line: 3, Col: 5},
		{Type: NL, Literal: "", Offset: 19, Line: 3, Col: 10},

		{Type: IDENT, Literal: "s", Offset: 20, Line: 4, Col: 1},
		{Type: WS, Literal: "", Offset: 21, Line: 4, Col: 2},
		{Type: ASSIGN, Literal: "", Offset: 22, Line: 4, Col: 3},
		{Type: WS, Literal: "", Offset: 23, Line: 4, Col: 4},
		{Type: IDENT, Literal: "s", Offset: 24, Line: 4, Col: 5},
		{Type: LBRACKET, Literal: "", Offset: 25, Line: 4, Col: 6},
		{Type: COLON, Literal: "", Offset: 26, Line: 4, Col: 7},
		{Type: NUM_LIT, Literal: "1", Offset: 27, Line: 4, Col: 8},
		{Type: RBRACKET, Literal: "", Offset: 28, Line: 4, Col: 9},
		{Type: WS, Literal: "", Offset: 29, Line: 4, Col: 10},
		{Type: COMMENT, Literal: `// "bc"`, Offset: 30, Line: 4, Col: 11},
		{Type: EOF, Literal: "", Offset: len(in), Line: 4, Col: 18},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, want, got)
	}
}

func TestMapRange(t *testing.T) {
	in := `
m := {name:"Mali" sport:"climbing"}
for key := range m
    print key m[key] // Mali climbing
end
`
	wantTokens := []*Token{
		{Type: NL, Literal: "", Offset: 0, Line: 1, Col: 1},
		{Type: IDENT, Literal: "m", Offset: 1, Line: 2, Col: 1},
		{Type: WS, Literal: "", Offset: 2, Line: 2, Col: 2},
		{Type: DECLARE, Literal: "", Offset: 3, Line: 2, Col: 3},
		{Type: WS, Literal: "", Offset: 5, Line: 2, Col: 5},
		{Type: LCURLY, Literal: "", Offset: 6, Line: 2, Col: 6},
		{Type: IDENT, Literal: "name", Offset: 7, Line: 2, Col: 7},
		{Type: COLON, Literal: "", Offset: 11, Line: 2, Col: 11},
		{Type: STRING_LIT, Literal: "Mali", Offset: 12, Line: 2, Col: 12},
		{Type: WS, Literal: "", Offset: 18, Line: 2, Col: 18},
		{Type: IDENT, Literal: "sport", Offset: 19, Line: 2, Col: 19},
		{Type: COLON, Literal: "", Offset: 24, Line: 2, Col: 24},
		{Type: STRING_LIT, Literal: "climbing", Offset: 25, Line: 2, Col: 25},
		{Type: RCURLY, Literal: "", Offset: 35, Line: 2, Col: 35},
		{Type: NL, Literal: "", Offset: 36, Line: 2, Col: 36},

		{Type: FOR, Literal: "", Offset: 37, Line: 3, Col: 1},
		{Type: WS, Literal: "", Offset: 40, Line: 3, Col: 4},
		{Type: IDENT, Literal: "key", Offset: 41, Line: 3, Col: 5},
		{Type: WS, Literal: "", Offset: 44, Line: 3, Col: 8},
		{Type: DECLARE, Literal: "", Offset: 45, Line: 3, Col: 9},
		{Type: WS, Literal: "", Offset: 47, Line: 3, Col: 11},
		{Type: RANGE, Literal: "", Offset: 48, Line: 3, Col: 12},
		{Type: WS, Literal: "", Offset: 53, Line: 3, Col: 17},
		{Type: IDENT, Literal: "m", Offset: 54, Line: 3, Col: 18},
		{Type: NL, Literal: "", Offset: 55, Line: 3, Col: 19},

		{Type: WS, Literal: "", Offset: 56, Line: 4, Col: 1},
		{Type: IDENT, Literal: "print", Offset: 60, Line: 4, Col: 5},
		{Type: WS, Literal: "", Offset: 65, Line: 4, Col: 10},
		{Type: IDENT, Literal: "key", Offset: 66, Line: 4, Col: 11},
		{Type: WS, Literal: "", Offset: 69, Line: 4, Col: 14},
		{Type: IDENT, Literal: "m", Offset: 70, Line: 4, Col: 15},
		{Type: LBRACKET, Literal: "", Offset: 71, Line: 4, Col: 16},
		{Type: IDENT, Literal: "key", Offset: 72, Line: 4, Col: 17},
		{Type: RBRACKET, Literal: "", Offset: 75, Line: 4, Col: 20},
		{Type: WS, Literal: "", Offset: 76, Line: 4, Col: 21},
		{Type: COMMENT, Literal: "// Mali climbing", Offset: 77, Line: 4, Col: 22},
		{Type: NL, Literal: "", Offset: 93, Line: 4, Col: 38},

		{Type: END, Literal: "", Offset: 94, Line: 5, Col: 1},
		{Type: NL, Literal: "", Offset: 97, Line: 5, Col: 4},

		{Type: EOF, Literal: "", Offset: len(in), Line: 6, Col: 1},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, want, got)
	}
}

func TestIllegal(t *testing.T) {
	tests := []struct {
		in   string
		want Token
	}{
		{in: ",  ", want: Token{Offset: 0, Literal: ",", Line: 1, Col: 1}},
		{in: `"unterminated`, want: Token{Offset: 0, Literal: "invalid string", Line: 1, Col: 1}},
		{in: `"newline in the
		middle "`, want: Token{Offset: 0, Literal: "invalid string", Line: 1, Col: 1}},
		{in: `"bad escape: \X"`, want: Token{Offset: 0, Literal: "invalid string", Line: 1, Col: 1}},
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
		{in: "\n", want: "NL\n"},
		{in: ` "abc"`, want: "WS\n" + `STRING_LIT "abc"` + "\n"},
		{in: ",", want: `ILLEGAL 💥 "," at line 1 column 1` + "\n"},
		{in: `"asdf `, want: `ILLEGAL 💥 "invalid string" at line 1 column 1` + "\n"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			got := run(tt.in)
			want := tt.want + "EOF\n"
			assert.Equal(t, want, got)
		})
	}
}

func TestRunWS(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: ":= ", want: "DECLARE\n"},
		{in: "  ", want: ""},
		{in: " \n ", want: "WS\nNL\n"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			got := run(tt.in)
			want := tt.want + "WS\nEOF\n"
			assert.Equal(t, want, got)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: `"abc" `, want: `STRING_LIT "abc"`},
		{in: `"abc\"" `, want: `STRING_LIT "abc\""`},
		{in: `"abc\\" `, want: `STRING_LIT "abc\\"`},
		{in: `"abc\\\"" `, want: `STRING_LIT "abc\\\""`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			got := run(tt.in)
			want := tt.want + "\nWS\nEOF\n"
			assert.Equal(t, want, got)
		})
	}
}

func TestStringLit(t *testing.T) {
	in := `fn "" 0`
	wantTokens := []*Token{
		{Type: IDENT, Literal: "fn", Offset: 0, Line: 1, Col: 1},
		{Type: WS, Literal: "", Offset: 2, Line: 1, Col: 3},
		{Type: STRING_LIT, Literal: "", Offset: 3, Line: 1, Col: 4},
		{Type: WS, Literal: "", Offset: 5, Line: 1, Col: 6},
		{Type: NUM_LIT, Literal: "0", Offset: 6, Line: 1, Col: 7},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, want, got)
	}
}

func TestStringLitErr(t *testing.T) {
	in := `fn "unterminated
"hello"`
	wantTokens := []*Token{
		{Type: IDENT, Literal: "fn", Offset: 0, Line: 1, Col: 1},
		{Type: WS, Literal: "", Offset: 2, Line: 1, Col: 3},
		{Type: ILLEGAL, Literal: "invalid string", Offset: 3, Line: 1, Col: 4},
		{Type: NL, Literal: "", Offset: 16, Line: 1, Col: 17},
		{Type: STRING_LIT, Literal: "hello", Offset: 17, Line: 2, Col: 1},
		{Type: EOF, Offset: 24, Line: 2, Col: 8},
	}
	l := New(in)
	for _, want := range wantTokens {
		got := l.Next()
		assert.Equal(t, want, got)
	}
}

func run(input string) string {
	l := New(input)
	tok := l.Next()
	var sb strings.Builder
	for ; tok.Type != EOF; tok = l.Next() {
		sb.WriteString(tok.String() + "\n")
	}
	sb.WriteString(tok.String() + "\n")
	return sb.String()
}
