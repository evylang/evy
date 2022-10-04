package parser

import (
	"strings"
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestParseDeclaration(t *testing.T) {
	tests := map[string][]string{
		"a := 1":     []string{"a=1"},
		"b:bool":     []string{"b=false"},
		"\nb:bool\n": []string{"b=false"},
		`a := "abc"
		b:bool
		c := true`: []string{"a='abc'", "b=false", "c=true"},
		"a:num[]":                      []string{"a=[]"},
		"a:num[]{}":                    []string{"a={}"},
		"abc:any[]{}":                  []string{"abc={}"},
		"a := bool[true]":              []string{"a=[true]"},
		"a := num[]":                   []string{"a=[]"},
		"a := num[][num[1 2]num[3 4]]": []string{"a=[[1, 2], [3, 4]]"},
		"a := num{a:1 b:2}":            []string{"a={a:1, b:2}"},
		"a := num[]{digits: num[1 2 3] nums: num[4 5]}": []string{"a={digits:[1, 2, 3], nums:[4, 5]}"},
		"a := num[]{digits: num[] nums: num[4]}":        []string{"a={digits:[], nums:[4]}"},
		"a := num[]{digits: num[4] nums: num[]}":        []string{"a={digits:[4], nums:[]}"},
		"a := num{}[]":                                  []string{"a=[]"},
		"a := num{}[num{}]":                             []string{"a=[{}]"},
		"a := any{a:1 b:true}":                          []string{"a={a:1, b:true}"},
		"a := any{a:1 b:true c:num[1]}":                 []string{"a={a:1, b:true, c:[1]}"},
		"a := num{}[num{a:1}]":                          []string{"a=[{a:1}]"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input, testBuiltins())
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
		parser := New(input, testBuiltins())
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
		"a := num{}[{a:1}]": "line 1 column 12: unexpected '{'", // TODO: expected `num` found `{`
		"a := num[true]":    "line 1 column 15: array literal 'true' should have type 'num'",
		"a := num{a:true}":  "line 1 column 16: map literal 'true' should have type 'num'",
		"a := num{}{":       "line 1 column 12: unterminated map literal",
		"a :=:":             "line 1 column 5: unexpected ':'",
		"a := num{":         "line 1 column 10: unterminated map literal",
		"a := num{}[":       "line 1 column 12: unterminated array literal",
		"a :num num":        "line 1 column 8: expected end of line, found 'num'",
		"a :num{}num":       "line 1 column 9: expected end of line, found 'num'",
	}
	for input, err1 := range tests {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, err1, parser.MaxErrorsString(1), "input: %s\nerrors:\n%s", input, parser.ErrorsString())
	}
}

func TestFunctionCall(t *testing.T) {
	tests := map[string][]string{
		"print":               []string{"print()"},
		"print 123":           []string{"print(123)"},
		`print 123 "abc"`:     []string{"print(123, 'abc')"},
		"a:=1 \n print a":     []string{"a=1", "print(a)"},
		`a := len "abc"`:      []string{"a=len('abc')"},
		`len "abc"`:           []string{"len('abc')"},
		`len num[]`:           []string{"len([])"},
		"a:string \n print a": []string{"a=''", "print(a)"},
		`a:=true
		b:string
		print a b`: []string{"a=true", "b=''", "print(a, b)"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input, testBuiltins())
		got := parser.Parse()
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestFunctionCallError(t *testing.T) {
	builtins := testBuiltins()
	builtins["f0"] = &FuncDecl{Name: "f0", ReturnType: NONE_TYPE}
	builtins["f1"] = &FuncDecl{Name: "f1", VariadicParam: &Var{Name: "a", T: NUM_TYPE}, ReturnType: NONE_TYPE}
	builtins["f2"] = &FuncDecl{Name: "f2", Params: []*Var{{Name: "a", T: NUM_TYPE}}, ReturnType: NONE_TYPE}
	builtins["f3"] = &FuncDecl{
		Name:       "f3",
		Params:     []*Var{{Name: "a", T: NUM_TYPE}, {Name: "b", T: STRING_TYPE}},
		ReturnType: NONE_TYPE,
	}
	tests := map[string]string{
		`len 2 2`:    "line 1 column 8: 'len' takes 1 argument, found 2",
		`len`:        "line 1 column 4: 'len' takes 1 argument, found 0",
		`a := print`: "line 1 column 11: invalid declaration, function 'print' has no return value",
		`a := f0`:    "line 1 column 8: invalid declaration, function 'f0' has no return value",
		`f0 "arg"`:   "line 1 column 9: 'f0' takes 0 arguments, found 1",
		`f2`:         "line 1 column 3: 'f2' takes 1 argument, found 0",
		`f2 f1`:      "line 1 column 4: function call must be parenthesized: (f1 ...)",
		`f1 "arg"`:   "line 1 column 9: 'f1' takes variadic arguments of type 'num', found 'string'",
		`f3 1 2`:     "line 1 column 7: 'f3' takes 2nd argument of type 'string', found 'num'",
		`f3 "1" "2"`: "line 1 column 11: 'f3' takes 1st argument of type 'num', found 'string'",
		`foo 0`:      "line 1 column 1: unknown function 'foo'",
	}
	for input, err1 := range tests {
		parser := New(input, builtins)
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, err1, parser.MaxErrorsString(1), "input: %s\nerrors:\n%s", input, parser.ErrorsString())
	}
}

func TestBlock(t *testing.T) {
	tests := map[string]string{
		`
if true
	print "TRUE"
end`: `
if (true) {
print('TRUE')
}
`[1:],
		`
if true
	if true
		print "TRUE"
	end
end`: `
if (true) {
if (true) {
print('TRUE')
}
}
`[1:],
	}
	for input, want := range tests {
		parser := New(input, testBuiltins())
		got := parser.Parse()
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestToplevelExprFuncCall(t *testing.T) {
	input := `
x := len "123"
`
	parser := New(input, testBuiltins())
	got := parser.Parse()
	assertNoParseError(t, parser, input)
	want := `
x=len('123')
`[1:]
	assert.Equal(t, want, got.String())
}

func TestFuncDecl(t *testing.T) {
	input := `
c := 1
func nums1:num n1:num n2:num
	if true //TODO: > 10
	    print c
	    return n1
	end
	return n2
end
on mousedown
	if true //TODO: > 10
	    print c
	end
end
func nums2:num n1:num n2:num
	if true //TODO: > 10
		return n1
	else
		return n2
	end
end
func nums3
	if true
		return
	end
end
return "success"
`
	parser := New(input, testBuiltins())
	_ = parser.Parse()
	assertNoParseError(t, parser, input)
	builtinCnt := len(testBuiltins())
	assert.Equal(t, builtinCnt+3, len(parser.funcs))
	got := parser.funcs["nums1"]
	assert.Equal(t, "nums1", got.Name)
	assert.Equal(t, NUM_TYPE, got.ReturnType)
	var wantVariadicParam *Var = nil
	assert.Equal(t, wantVariadicParam, got.VariadicParam)
	assert.Equal(t, 2, len(got.Params))
	n1 := got.Params[0]
	assert.Equal(t, "n1", n1.Name)
	assert.Equal(t, NUM_TYPE, n1.Type())
	assert.Equal(t, 2, len(got.Body.Statements))
	returnStmt := got.Body.Statements[1]
	assert.Equal(t, "return n2", returnStmt.String())
}

func TestReturnErr(t *testing.T) {
	inputs := map[string]string{`
func add:num
	return 1
	print "boom"
end
`: "line 4 column 2: unreachable code",
		`
func nums:num
	if true
		return 1
	else
		return 2
	end
	print "boom"
end
`: "line 8 column 2: unreachable code",
		`
func nums:num
	if true
		if true
			return 3
		else
			return 4
		end
	else
		return 2
	end
	print "boom"
end
`: "line 12 column 2: unreachable code",
		`
func nums:num
	a := 5
	while true
		return 1
	end
	print "boom"
end
`: "line 7 column 2: unreachable code",
		`
foo
return false
func foo
  print "hello"
end
print "do i run?"
`: "line 7 column 1: unreachable code",
	}
	for input, wantErr := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, wantErr, parser.MaxErrorsString(1))
	}
}

func TestFuncAssignment(t *testing.T) {
	inputs := []string{`
a := 1
b:num
b = a
`, `
a:num
b:num
b = a
`, `
a:num
b:any
b = a
`,
	}
	for _, input := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertNoParseError(t, parser, input)
	}
}

func TestFuncAssignmentErr(t *testing.T) {
	inputs := map[string]string{`
b:num
b = true
`: "line 3 column 3: 'b' accepts values of type num, found bool",
		`
a:= 1
a = b
`: "line 3 column 6: unknown variable name 'b'",
		`
a:= 1
b = a
`: "line 3 column 3: unknown variable name 'b'",
		`
a:= 1
a = num[]
`: "line 3 column 3: 'a' accepts values of type num, found num[]",
		`
a:num
b:any
a = b
`: "line 4 column 3: 'a' accepts values of type num, found any",
		`
func fn:bool
	return true
end
fn = 3
`: "line 5 column 1: cannot assign to 'fn' as it is a function not a variable",
	}
	for input, wantErr := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, wantErr, parser.MaxErrorsString(1))
	}
}

func TestScope(t *testing.T) {
	inputs := []string{`
x := 1
func foo
	x := "abc"
end
`, `
x := 1
func foo x:string
	x = "abc"
end
`, `
x := 1
func foo x:string...
	print x
end
`, `
x := 1
if true
	x := "abc" // block scope
end
`,
	}
	for _, input := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertNoParseError(t, parser, input)
	}
}

func TestScopeErr(t *testing.T) {
	inputs := map[string]string{
		`
x := 1
x := 2
`: "line 3 column 1: redeclaration of 'x'",
		`
x := 1
x := "abc"
`: "line 3 column 1: redeclaration of 'x'",
		`
x :num
x := "abc"
`: "line 3 column 1: redeclaration of 'x'",
		`
x := "abc"
x :num
`: "line 3 column 1: redeclaration of 'x'",
		`
x :num
x :num
`: "line 3 column 1: redeclaration of 'x'",
		`
x :num
x :string
`: "line 3 column 1: redeclaration of 'x'",
		`
x :num
func x
   print "abc"
end
`: "line 2 column 1: invalid declaration of 'x', already used as function name",
		`
func x in:num
   in:string
end
`: "line 3 column 4: redeclaration of 'in'",
		`
func foo
   x := 0
   x := 0
end
`: "line 4 column 4: redeclaration of 'x'",
		`
func x
   x := 0
end
`: "line 3 column 4: invalid declaration of 'x', already used as function name",
		`
func x in:string in:string
   print in
end
`: "line 2 column 18: redeclaration of parameter 'in'",
		`
func x x:string
   print x
end
`: "line 2 column 8: invalid declaration of parameter 'x', already used as function name",
		`
func x x:string...
   print x
end
`: "line 2 column 8: invalid declaration of parameter 'x', already used as function name",
	}
	for input, wantErr := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, wantErr, parser.MaxErrorsString(1))
	}
}

func TestIf(t *testing.T) {
	inputs := []string{
		`if true
			print "yeah"
		end`,
		`if true
			print "true"
		 else
			print "false"
		 end`,
		`if true
			print "true"
		 else if false
			print "false"
		 end`,
		`if true
			print "true"
		 else if false
			print "false"
		 else if true
			print "true true"
		 else
			print "false"
		 end`,
		`if true
			if true
				print "true true"
			else
				print "true false"
			end
		 else
			if true
				print "false true"
			else
				print "false false"
			end
		 end`,
	}
	for _, input := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertNoParseError(t, parser, input)
	}
}

func TestIfErr(t *testing.T) {
	inputs := map[string]string{
		`
if true
	print "baba yaga"
`: "line 4 column 1: expected 'end', got end of input",
		`
if true
end`: "line 3 column 1: at least one statement is required here",
		`
if
	print "baba yaga"
end`: "line 2 column 3: unexpected end of line",
		`
if true
	print "true"
else true
	print "true"
end`: "line 4 column 6: expected end of line, found 'true'",
		`
if true
	print "true"
else if
	print "true"
end`: "line 4 column 8: unexpected end of line",
		`
if true
	print "true"
else
   print "false"
else if false
	print "true"
end`: "line 6 column 1: unexpected input 'else'",
		`
if true
	if true
		print "true true"
else
	print "false"
end`: "line 7 column 4: expected 'end', got end of input",
	}
	for input, wantErr := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, wantErr, parser.MaxErrorsString(1), "input: %s", input)
	}
}

func TestWhile(t *testing.T) {
	inputs := []string{
		`
while true
	print "forever"
end`,
		`
while has_more
	print "🍭"
end

two_more := true
one_more := true
func has_more:bool
	if one_more
		if two_more
			two_more = false
			return false
		else
			one_more = false
			return false
		end
	end
	return true
end
`,
	}
	for _, input := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertNoParseError(t, parser, input)
	}
}

func TestWhileErr(t *testing.T) {
	inputs := map[string]string{
		`
while true
	print "forever"
`: "line 4 column 1: expected 'end', got end of input",
		`
while true
end`: "line 3 column 1: at least one statement is required here",
		`
while
	print "forever"
end`: "line 2 column 6: unexpected end of line",
	}
	for input, wantErr := range inputs {
		parser := New(input, testBuiltins())
		_ = parser.Parse()
		assertParseError(t, parser, input)
		assert.Equal(t, wantErr, parser.MaxErrorsString(1), "input: %s", input)
	}
}

func TestDemo(t *testing.T) {
	input := `
move 10 10
line 20 20

x := 12
print "x:" x
if true //TODO: x > 10
    print "🍦 big x"
end`
	parser := New(input, testBuiltins())
	got := parser.Parse()
	assertParseError(t, parser, input)
	assert.Equal(t, "line 2 column 1: unknown function 'move'", parser.MaxErrorsString(1))
	assert.Equal(t, "line 3 column 1: unknown function 'line'", parser.errors[1].String())
	want := `
x=12
print('x:', x)
if (true) {
print('🍦 big x')
}
`[1:]
	assert.Equal(t, want, got.String())
}

func assertParseError(t *testing.T, parser *Parser, input string) {
	t.Helper()
	assert.Equal(t, true, len(parser.errors) > 0, "expected parser errors, got none: input: %s\n", input)
}

func assertNoParseError(t *testing.T, parser *Parser, input string) {
	t.Helper()
	assert.Equal(t, 0, len(parser.errors), "Unexpected parser error\n input: %s\nerrors:\n%s", input, parser.ErrorsString())
}

func testBuiltins() map[string]*FuncDecl {
	return map[string]*FuncDecl{
		"print": &FuncDecl{
			Name:          "print",
			VariadicParam: &Var{Name: "a", T: ANY_TYPE},
			ReturnType:    NONE_TYPE,
		},
		"len": &FuncDecl{
			Name:       "len",
			Params:     []*Var{{Name: "a", T: ANY_TYPE}},
			ReturnType: NUM_TYPE,
		},
	}
}
