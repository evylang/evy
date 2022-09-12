package parser

import (
	"strings"
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestParseDeclaration(t *testing.T) {
	tests := map[string][]string{
		"a := 1":     []string{"a:NUM=1"},
		"b:bool":     []string{"b:BOOL=false"},
		"\nb:bool\n": []string{"b:BOOL=false"},
		`a := "abc"
		b:bool
		c := true`: []string{"a:STRING='abc'", "b:BOOL=false", "c:BOOL=true"},
		"a:num[]":                      []string{"a:ARRAY NUM=[]"},
		"a:num[]{}":                    []string{"a:MAP ARRAY NUM={}"},
		"abc:any[]{}":                  []string{"abc:MAP ARRAY ANY={}"},
		"a := bool[true]":              []string{"a:ARRAY BOOL=[true]"}, // TODO: should be print "a:ARRAY BOOL=[true]
		"a := num[]":                   []string{"a:ARRAY NUM=[]"},
		"a := num[][num[1 2]num[3 4]]": []string{"a:ARRAY ARRAY NUM=[[1, 2], [3, 4]]"},
		"a := num{a:1 b:2}":            []string{"a:MAP NUM={a:1, b:2}"},
		"a := num[]{digits: num[1 2 3] nums: num[4 5]}": []string{"a:MAP ARRAY NUM={digits:[1, 2, 3], nums:[4, 5]}"},
		"a := num[]{digits: num[] nums: num[4]}":        []string{"a:MAP ARRAY NUM={digits:[], nums:[4]}"},
		"a := num[]{digits: num[4] nums: num[]}":        []string{"a:MAP ARRAY NUM={digits:[4], nums:[]}"},
		"a := num{}[]":                                  []string{"a:ARRAY MAP NUM=[]"},
		"a := num{}[num{}]":                             []string{"a:ARRAY MAP NUM=[{}]"},
		"a := any{a:1 b:true}":                          []string{"a:MAP ANY={a:1, b:true}"},
		"a := any{a:1 b:true c:num[1]}":                 []string{"a:MAP ANY={a:1, b:true, c:[1]}"},
		"a := num{}[num{a:1}]":                          []string{"a:ARRAY MAP NUM=[{a:1}]"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input)
		got := parser.Parse()
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestEmptyProgram(t *testing.T) {
	tests := []string{
		"",
		"\n",
		"\n\n\n",
		" ",
		" \n //adf \n",
		"//blabla",
		"//blabla\n",
		" \n //blabla \n",
		" \n //blabla",
	}
	for _, input := range tests {
		parser := New(input)
		got := parser.Parse()
		assertNoParseError(t, parser, input)
		assert.Equal(t, "\n", got.String())
	}
}

func TestParseDeclarationError(t *testing.T) {
	tests := map[string]string{
		"a :invalid":        "line 1 column 1: invalid type declaration for 'a'",
		"a :":               "line 1 column 1: invalid type declaration for 'a'",
		"a :\n":             "line 1 column 1: invalid type declaration for 'a'",
		"a ://blabla\n":     "line 1 column 1: invalid type declaration for 'a'",
		"a :true":           "line 1 column 1: invalid type declaration for 'a'",
		"a :[]":             "line 1 column 1: invalid type declaration for 'a'",
		"a :[]num":          "line 1 column 1: invalid type declaration for 'a'",
		"a :()":             "line 1 column 1: invalid type declaration for 'a'",
		"a ::":              "line 1 column 1: invalid type declaration for 'a'",
		"a := num{}[{a:1}]": "line 1 column 12: unexpected character '{'", // TODO: expected `num` found `{`
		"a := num[true]":    "line 1 column 15: array literal 'true' should have type 'num'",
		"a := num{a:true}":  "line 1 column 16: map literal 'true' should have type 'num'",
		"a := num{}{":       "line 1 column 12: unterminated map literal",
		"a :=:":             "line 1 column 5: unexpected character ':'",
		"a := num{":         "line 1 column 10: unterminated map literal",
		"a := num{}[":       "line 1 column 12: unterminated array literal",
		"a :num num":        "line 1 column 8: expected end of line, found 'num'",
		"a :num{}num":       "line 1 column 9: expected end of line, found 'num'",
	}
	for input, err1 := range tests {
		parser := New(input)
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, err1, parser.errors[0].String(), "input: %s\nerrors:\n%s", input, parser.errorsString())
	}
}

func TestFunctionCall(t *testing.T) {
	tests := map[string][]string{
		"print":               []string{"print()"},
		"print 123":           []string{"print(123)"},
		`print 123 "abc"`:     []string{"print(123, 'abc')"},
		"a:=1 \n print a":     []string{"a:NUM=1", "print(a:NUM)"},
		`a := len "abc"`:      []string{"a:NUM=len('abc')"},
		`len "abc"`:           []string{"len('abc')"},
		`len num[]`:           []string{"len([])"},
		"a:string \n print a": []string{"a:STRING=''", "print(a:STRING)"},
		`a:=true
		b:string
		print a b`: []string{"a:BOOL=true", "b:STRING=''", "print(a:BOOL, b:STRING)"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input)
		got := parser.Parse()
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestFunctionCallError(t *testing.T) {
	builtins := builtins()
	builtins["f0"] = &FuncDecl{Name: "f0", ReturnType: NONE_TYPE}
	builtins["f1"] = &FuncDecl{Name: "f1", VariadicParam: &Var{Name: "a", nType: NUM_TYPE}, ReturnType: NONE_TYPE}
	builtins["f2"] = &FuncDecl{Name: "f2", Params: []*Var{&Var{Name: "a", nType: NUM_TYPE}}, ReturnType: NONE_TYPE}
	builtins["f3"] = &FuncDecl{
		Name:       "f3",
		Params:     []*Var{&Var{Name: "a", nType: NUM_TYPE}, &Var{Name: "b", nType: STRING_TYPE}},
		ReturnType: NONE_TYPE,
	}
	tests := map[string]string{
		`len 2 2`:    "line 1 column 8: 'len' takes 1 argument, found 2",
		`len`:        "line 1 column 4: 'len' takes 1 argument, found 0",
		`a := print`: "line 1 column 11: invalid declaration, function 'print' has no return value",
		`a := f0`:    "line 1 column 8: invalid declaration, function 'f0' has no return value",
		`f0 "arg"`:   "line 1 column 9: 'f0' takes 0 arguments, found 1",
		`f2`:         "line 1 column 3: 'f2' takes 1 argument, found 0",
		`f1 "arg"`:   "line 1 column 9: 'f1' takes variadic arguments of type 'num', found 'string'",
		`f3 1 2`:     "line 1 column 7: 'f3' takes 2nd argument of type 'string', found 'num'",
		`f3 "1" "2"`: "line 1 column 11: 'f3' takes 1st argument of type 'num', found 'string'",
		`foo 0`:      "line 1 column 1: unknown function 'foo'",
	}
	for input, err1 := range tests {
		parser := NewWithBuiltins(input, builtins)
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, err1, parser.errors[0].String(), "input: %s\nerrors:\n%s", input, parser.errorsString())
	}
}

func TestBlock(t *testing.T) {
	tests := map[string][]string{
		`if true
			print "TRUE"
		end`: []string{""},
		`if true
			if 12 > 11
				print "TRUE"
			end
		end`: []string{""},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input)
		got := parser.Parse()
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestToplevelExprFuncCall(t *testing.T) {
	input := `
x := len "123"
`
	parser := New(input)
	got := parser.Parse()
	assertNoParseError(t, parser, input)
	want := `
x:NUM=len('123')
`[1:]
	assert.Equal(t, want, got.String())
}

func TestFuncDecl(t *testing.T) {
	input := `
c := add 1 2
func add:num n1:num n2:num
	if c > 10
	    print c
	end
	return n1 + n2
end
on mousedown
	if c > 10
	    print c
	end
end
`
	parser := New(input)
	_ = parser.Parse()
	assertNoParseError(t, parser, input)
	builtinCnt := len(builtins())
	assert.Equal(t, builtinCnt+1, len(parser.funcs))
	got := parser.funcs["add"]
	assert.Equal(t, "add", got.Name)
	assert.Equal(t, NUM_TYPE, got.ReturnType)
	var wantVariadicParam *Var = nil
	assert.Equal(t, wantVariadicParam, got.VariadicParam)
	assert.Equal(t, 2, len(got.Params))
	n1 := got.Params[0]
	assert.Equal(t, "n1", n1.Name)
	assert.Equal(t, NUM_TYPE, n1.Type())
	assert.Equal(t, 0, len(got.Body.Statements))
}

func TestDemo(t *testing.T) {
	input := `
move 10 10
line 20 20

x := 12
print "x:" x
if x > 10
    print "🍦 big x"
end`
	parser := New(input)
	got := parser.Parse()
	assertParseError(t, parser, input)
	assert.Equal(t, "line 2 column 1: unknown function 'move'", parser.errors[0].String())
	assert.Equal(t, "line 3 column 1: unknown function 'line'", parser.errors[1].String())
	want := `
x:NUM=12
print('x:', x:NUM)
`[1:]
	assert.Equal(t, want, got.String())
}

func assertParseError(t *testing.T, parser *Parser, input string) {
	t.Helper()
	assert.Equal(t, true, len(parser.errors) > 0, "expected parser errors, got none: input: %s\n", input)
}

func assertNoParseError(t *testing.T, parser *Parser, input string) {
	t.Helper()
	assert.Equal(t, 0, len(parser.errors), "Unexpected parser error\n input: %s\nerrors:\n%s", input, parser.errorsString())
}
