package parser

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
)

func TestParseTopLevelExpression(t *testing.T) {
	tests := map[string]string{
		// literals and variables
		"1":   "1",
		"n1":  "n1",
		"n2":  "n2",
		"s":   "s",
		"b":   "b",
		`"b"`: `"b"`,

		// binary expressions, arithmetic
		"1+1":   "(1+1)",
		"1* n1": "(1*n1)",
		"1*2%3": "((1*2)%3)",
		"1*2/3": "((1*2)/3)",
		"1+2*3": "(1+(2*3))",
		"n1/n2": "(n1/n2)",

		// binary expressions, logical
		"n1<n2":                "(n1<n2)",
		"1<2":                  "(1<2)",
		`"a"<"b"`:              `("a"<"b")`,
		`s<"b"`:                `(s<"b")`,
		"n1== n2":              "(n1==n2)",
		"false and true":       "(false and true)",
		"false or b":           "(false or b)",
		"false or b and false": "(false or (b and false))",

		// binary expressions, combined
		"n1<n2 or n1== n2":      "((n1<n2) or (n1==n2))",
		"1<n2 or n1== n2 and b": "((1<n2) or ((n1==n2) and b))",

		// grouped expressions
		"1+(2-3)":                "(1+(2-3))",
		"(1+2)+3":                "((1+2)+3)",
		"(1*2)+3":                "((1*2)+3)",
		"(1+2)*3":                "((1+2)*3)",
		"n1<3 + 7 * (n2-1)":      "(n1<(3+(7*(n2-1))))",
		"n1<3 + 7 * (n2-1)or !b": "((n1<(3+(7*(n2-1)))) or (!b))",

		// unary expressions
		"-5":    "(-5)",
		"-n1":   "(-n1)",
		"!true": "(!true)",
		"!b":    "(!b)",

		// function calls
		"print":        "print()",
		"print 1":      "print(any(1))",
		"print 1 true": "print(any(1), any(true))",
		`print 1 "a"`:  `print(any(1), any("a"))`,

		// Function calls
		`print (1+2)`:                 "print(any((1+2)))",
		`print 1+2`:                   "print(any((1+2)))",
		`print 1+2 3+4`:               "print(any((1+2)), any((3+4)))",
		`print 1-2`:                   "print(any((1-2)))",
		`print 1 -2`:                  "print(any(1), any((-2)))",
		`len "abc"`:                   `len(any("abc"))`,
		`print (len "abc")`:           `print(any(len(any("abc"))))`,
		`print 1 (len "abc")`:         `print(any(1), any(len(any("abc"))))`,
		`print 1 (len "abc") 2`:       `print(any(1), any(len(any("abc"))), any(2))`,
		`print (len "abc") 2`:         `print(any(len(any("abc"))), any(2))`,
		`print (len "abc") (len "x")`: `print(any(len(any("abc"))), any(len(any("x"))))`,
		`print s[1]`:                  "print(any((s[1])))",
		"print map2[s]":               "print(any((map2[s])))",

		// niladic
		`print rand1`:       "print(any(rand1()))",
		`print rand1 rand1`: "print(any(rand1()), any(rand1()))",

		// // Index expression
		"arr[1]":        "(arr[1])",
		"arr2[1][2]":    "((arr2[1])[2])",
		"arr2[1][n2]":   "((arr2[1])[n2])",
		"arr2[1][n2+2]": "((arr2[1])[(n2+2)])",
		"arr[1] and b":  "((arr[1]) and b)",
		"map[s]":        "(map[s])",
		`map["key"]`:    `(map["key"])`,
		`"abc"[1]`:      `("abc"[1])`,
		`s[1]`:          "(s[1])",

		// Map access - dot expressions
		"map.key":          "(map.key)",
		"map.end":          "(map.end)",
		"map.key+3":        "((map.key)+3)",
		"map2.a.b":         "((map2.a).b)",
		"map.key+map2.a.b": "((map.key)+((map2.a).b))",
		"map3.ok[1]":       "((map3.ok)[1])",
		"map3.ok[n1]":      "((map3.ok)[n1])",
		"list[1].x":        "((list[1]).x)",
		"list[n1][s]":      "((list[n1])[s])",
		"map2[s]":          "(map2[s])",

		// Type assertions
		"a.(num)":     "(a.(num))",
		"a.(string)":  "(a.(string))",
		"a.(bool)":    "(a.(bool))",
		"a.([]num)":   "(a.([]num))",
		"a.({}[]num)": "(a.({}[]num))",

		// Array literals
		"[]":          "[]",
		"[1]":         "[1]",
		"[1 2]":       "[1, 2]",
		"[n1]":        "[n1]",
		"[n1 n2]":     "[n1, n2]",
		"[n1 2]":      "[n1, 2]",
		"[n1 1 n2 2]": "[n1, 1, n2, 2]",
		"[(n1+1)]":    "[(n1+1)]",
		"[(n1+1) 2]":  "[(n1+1), 2]",
		"[(1)]":       "[1]",

		// Combined array literals
		"[[] 1]":             "[any([]), any(1)]",
		"[ [] 1 ]":           "[any([]), any(1)]",
		"[[] [1]]":           "[[], [1]]",
		"[[1 2][1]]":         "[([1, 2][1])]",
		"[[]]":               "[[]]",
		"[[1]]":              "[[1]]",
		"[[1] ([])]":         "[[1], []]",
		"[[] ([])]":          "[[], []]",
		"[[] ([1])]":         "[[], [1]]",
		"[[] 1 true n2]":     "[any([]), any(1), any(true), any(n2)]",
		"[1 2 3][1]":         "([1, 2, 3][1])",
		"[ 3+5 n1*2]":        "[(3+5), (n1*2)]",
		"len []":             "len(any([]))",
		"[ [] { a : 2+3 } ]": "[any([]), any({a:(2+3)})]",
		// Map literals
		"{}":                     "{}",
		"{a: 1}":                 "{a:1}",
		"{ a: 1 }":               "{a:1}",
		"{a: 1 b:2}":             "{a:1, b:2}",
		"{a: [1] b:2}":           "{a:any([1]), b:any(2)}",
		"{a: [1] b:2 c: 1+2}":    "{a:any([1]), b:any(2), c:any((1+2))}",
		"{a: [1] b:2+n2 c: 1+2}": "{a:any([1]), b:any((2+n2)), c:any((1+2))}",
		"{a: 1}.a":               "({a:1}.a)",
		`{a: 1}["a"]`:            `({a:1}["a"])`,
		`{end: 1}["end"]`:        `({end:1}["end"])`,

		// Array concatenation
		"[1] + [2]":            "([1]+[2])",
		"[true] + [false]":     "([true]+[false])",
		"[1 true] + [2 false]": "([any(1), any(true)]+[any(2), any(false)])",
		"[1] + []":             "([1]+[])",
		"[] + [1]":             "([]+[1])",
		"[] + []":              "([]+[])",
		"[[1]]+[[]]":           "([[1]]+[[]])",
		"[[]]+[[1]]":           "([[]]+[[1]])",

		// Array repetition
		"[1] * 2":  "([1]*2)",
		"[] * 0":   "([]*0)",
		"[1] * n1": "([1]*n1)",
		"arr * 2":  "(arr*2)",
		"arr * n1": "(arr*n1)",

		// Slices
		"arr[1:2]": "(arr[1:2])",
		"arr[1:]":  "(arr[1:])",
		"arr[:2]":  "(arr[:2])",
		"arr[:]":   "(arr[:])",

		// Multiline declarations
		`[ 1
		   2 ]`: "[1, 2]",
		"[  " + `
		   1

		   2
		 ]`: "[1, 2]",
		`[
		   1
		   2
		 ]`: "[1, 2]",
		`{ a:1
		   b:2 }`: "{a:1, b:2}",
		`{
		   a:1
		   b:2
		}`: "{a:1, b:2}",
		"{  " + `
		   a:1

		   b:2   ` + `

		}`: "{a:1, b:2}",
	}
	for input, want := range tests {
		parser := newParser(input, testBuiltins())
		parser.formatting = newFormatting()
		parser.advanceTo(0)
		parser.scope = newScope(nil, &Program{})
		parser.scope.set("n1", &Var{Name: "n1", T: NUM_TYPE})
		parser.scope.set("n2", &Var{Name: "n2", T: NUM_TYPE})
		parser.scope.set("s", &Var{Name: "s", T: STRING_TYPE})
		parser.scope.set("b", &Var{Name: "b", T: BOOL_TYPE})
		parser.scope.set("a", &Var{Name: "a", T: ANY_TYPE})
		arrType := &Type{Name: ARRAY, Sub: BOOL_TYPE}
		parser.scope.set("arr", &Var{Name: "arr", T: arrType})
		arrType2 := &Type{Name: ARRAY, Sub: arrType}
		parser.scope.set("arr2", &Var{Name: "arr2", T: arrType2})
		mapType := &Type{Name: MAP, Sub: NUM_TYPE}
		parser.scope.set("map", &Var{Name: "map", T: mapType})
		mapType2 := &Type{Name: MAP, Sub: mapType}
		parser.scope.set("map2", &Var{Name: "map2", T: mapType2})
		arrayMapType := &Type{Name: ARRAY, Sub: mapType}
		parser.scope.set("list", &Var{Name: "list", T: arrayMapType})
		mapArrayType := &Type{Name: MAP, Sub: arrType}
		parser.scope.set("map3", &Var{Name: "map3", T: mapArrayType})

		got := parser.parseTopLevelExpr()
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestParseTopLevelExpressionErr(t *testing.T) {
	tests := map[string]string{
		"x":        `line 1 column 1: unknown variable name "x"`,
		"+1":       `line 1 column 1: unexpected "+"`,
		"* n1":     `line 1 column 1: unexpected "*"`,
		"and true": `line 1 column 1: unexpected "and"`,

		"1 +":    "line 1 column 4: unexpected end of input",
		"1 +\n2": "line 1 column 4: unexpected end of line",
		"1 ==":   "line 1 column 5: unexpected end of input",

		"true + false": `line 1 column 6: "+" takes num, string or array type, found bool`,
		"true - false": `line 1 column 6: "-" takes num type, found bool`,
		"true < false": `line 1 column 6: "<" takes num or string type, found bool`,
		"1 and 2":      `line 1 column 3: "and" takes bool type, found num`,
		"1 + false":    `line 1 column 3: mismatched type for +: num, bool`,
		"false + 1":    `line 1 column 7: mismatched type for +: bool, num`,
		"[1] * true":   `line 1 column 5: array repetition ("*") takes num on right, found bool`,
		"(1+2":         `line 1 column 5: expected ")", got end of input`,
		"(1+2\n)":      `line 1 column 5: expected ")", got end of line`,
		"(1+)2":        `line 1 column 4: unexpected ")"`,
		"(1+]2":        `line 1 column 4: unexpected "]"`,
		"(1+2]":        `line 1 column 5: expected ")", got "]"`,

		`"abc"["a"]`:   "line 1 column 6: string index expects num, found string",
		`[1 2 3]["a"]`: "line 1 column 8: array index expects num, found string",
		"{a:2}[2]":     "line 1 column 6: map index expects string, found num",

		`"abc"["a":]`:    "line 1 column 6: string start index expects num, found string",
		`"abc"["a":"b"]`: "line 1 column 6: string start index expects num, found string",
		`"abc"[:"b"]`:    "line 1 column 6: string end index expects num, found string",

		"{a:}": `line 1 column 4: unexpected "}"`,
		"{:a}": `line 1 column 2: expected map key, found ":"`,

		"[1] + [false]": "line 1 column 5: mismatched type for +: []num, []bool",

		"n1.(num)": "line 1 column 3: value of type assertion must be of type any, not num",
		"a.(any)":  "line 1 column 2: cannot type assert to type any",
		"a.(x)":    `line 1 column 2: invalid type in type assertion of "a"`,
		"a.([]x)":  `line 1 column 2: invalid type in type assertion of "a"`,

		"a. (num)":    `line 1 column 2: unexpected whitespace after "."`,
		"a .(num)":    `line 1 column 3: unexpected whitespace before "."`,
		"map. b":      `line 1 column 4: unexpected whitespace after "."`,
		"map .b":      `line 1 column 5: unexpected whitespace before "."`,
		"- 5":         `line 1 column 1: unexpected whitespace after "-"`,
		"- n1":        `line 1 column 1: unexpected whitespace after "-"`,
		"[3 +5]":      `line 1 column 4: unexpected whitespace before "+"`,
		"[3+ 5]":      `line 1 column 3: unexpected whitespace after "+"`,
		"[ 3+ 5]":     `line 1 column 4: unexpected whitespace after "+"`,
		"print 1 - 2": `line 1 column 9: unexpected whitespace after "-"`,

		"- 2":    `line 1 column 1: unexpected whitespace after "-"`,
		"! true": `line 1 column 1: unexpected whitespace after "!"`,

		"{a: _}":   `line 1 column 5: anonymous variable "_" cannot be read`,
		"[_]":      `line 1 column 2: anonymous variable "_" cannot be read`,
		"{a:1}[_]": `line 1 column 7: anonymous variable "_" cannot be read`,

		"[1":    `line 1 column 3: expected "]", got end of input`,
		"[1)":   `line 1 column 3: unexpected ")"`,
		"[1(]":  `line 1 column 4: unexpected "]"`,
		"[1()]": `line 1 column 4: unexpected ")"`,
	}
	for input, wantErr := range tests {
		parser := newParser(input, testBuiltins())
		parser.advanceTo(0)
		parser.formatting = newFormatting()
		parser.scope = newScope(nil, &Program{})
		parser.scope.set("n1", &Var{Name: "n1", T: NUM_TYPE})
		mapType := &Type{Name: MAP, Sub: NUM_TYPE}
		parser.scope.set("map", &Var{Name: "map", T: mapType})
		parser.scope.set("a", &Var{Name: "a", T: ANY_TYPE})

		_ = parser.parseTopLevelExpr()
		assertParseError(t, parser, input)
		got := parser.errors.Truncate(1)
		assert.Equal(t, wantErr, got.Error(), "input: %s\nerrors:\n%s", input, parser.errors)
	}
}

func TestParseNiladic(t *testing.T) {
	inputs := map[string]string{
		"rand1":       "print rand1",
		"rand1-twice": "print rand1 rand1",
		"rand1-expr":  "print rand1+10",
		"rand1-toplevelexpr": `
n := rand1+0.5
print n
`,
		"rand1-toplevelexpr-space": `
if rand1 > 0.5
	print "big"
end
`,
		"custom1": `
func answer:string
	return "42"
end

print answer`,
	}

	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			parser := newParser(input, testBuiltins())
			_ = parser.parse()
			assertNoParseError(t, parser, input)
		})
	}
}

func TestParseNiladicErr(t *testing.T) {
	type inputWithError struct {
		input      string
		wantErrMsg string
	}
	inputs := map[string]inputWithError{
		"rand1-group": {
			input:      "print (rand1 100)",
			wantErrMsg: `line 1 column 14: expected ")", got ""`,
		},
		"custom1": {
			input: `
func answer
	print "42"
end

len answer`,
			wantErrMsg: `line 6 column 5: cannot use "answer" as 1st argument, it has no return value`,
		},
		"custom1-variadic": {
			input: `
func answer
	print "42"
end

print answer`,
			wantErrMsg: `line 6 column 7: cannot use "answer" as argument, it has no return value`,
		},
	}

	for name, tc := range inputs {
		t.Run(name, func(t *testing.T) {
			parser := newParser(tc.input, testBuiltins())
			_ = parser.parse()
			assertParseError(t, parser, tc.input)
			gotErr := parser.errors.Truncate(1)
			assert.Equal(t, tc.wantErrMsg, gotErr.Error())
		})
	}
}
