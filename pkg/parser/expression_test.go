package parser

import (
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestParseTopLevelExpression(t *testing.T) {
	tests := map[string]string{
		// literals and variables
		"1":   "1",
		"n1":  "n1",
		"n2":  "n2",
		"s":   "s",
		"b":   "b",
		`"b"`: "'b'",

		// binary expressions, arithmetic
		"1+1":   "(1+1)",
		"1* n1": "(1*n1)",
		"1*2*3": "((1*2)*3)",
		"1*2/3": "((1*2)/3)",
		"1+2*3": "(1+(2*3))",
		"n1/n2": "(n1/n2)",

		// binary expressions, logical
		"n1<n2":                "(n1<n2)",
		"1<2":                  "(1<2)",
		`"a"<"b"`:              `('a'<'b')`,
		`s<"b"`:                `(s<'b')`,
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
		"print 1":      "print(1)",
		"print 1 true": "print(1, true)",
		`print 1 "a"`:  "print(1, 'a')",

		// Function calls
		`print (1+2)`:                 "print((1+2))",
		`len "abc"`:                   "len('abc')",
		`print (len "abc")`:           "print(len('abc'))",
		`print 1 (len "abc")`:         "print(1, len('abc'))",
		`print 1 (len "abc") 2`:       "print(1, len('abc'), 2)",
		`print (len "abc") 2`:         "print(len('abc'), 2)",
		`print (len "abc") (len "x")`: "print(len('abc'), len('x'))",
		`print s[1]`:                  "print((s[1]))",

		// Index expression
		"arr[1]":        "(arr[1])",
		"arr2[1][2]":    "((arr2[1])[2])",
		"arr2[1][n2]":   "((arr2[1])[n2])",
		"arr2[1][n2+2]": "((arr2[1])[(n2+2)])",
		"arr[1] and b":  "((arr[1]) and b)",
		"map[s]":        "(map[s])",
		`map["key"]`:    "(map['key'])",
		`"abc"[1]`:      "('abc'[1])",
		`s[1]`:          "(s[1])",

		// Map access - dot expressions
		"map.key":          "(map.key)",
		"map.key+3":        "((map.key)+3)",
		"map2.a.b":         "((map2.a).b)",
		"map.key+map2.a.b": "((map.key)+((map2.a).b))",
		"map3.ok[1]":       "((map3.ok)[1])",
		"map3.ok[n1]":      "((map3.ok)[n1])",
		"list[1].x":        "((list[1]).x)",
		"list[n1][s]":      "((list[n1])[s])",

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
		"[[] 1]":         "[[], 1]",
		"[[] [1]]":       "[([][1])]",
		"[[]]":           "[[]]",
		"[[1]]":          "[[1]]",
		"[[1] ([])]":     "[[1], []]",
		"[[] ([])]":      "[[], []]",
		"[[] ([1])]":     "[[], [1]]",
		"[[] 1 true n2]": "[[], 1, true, n2]",
		"[1 2 3][1]":     "([1, 2, 3][1])",
		"len []":         "len([])",

		// Map literals
		"{}":                     "{}",
		"{a: 1}":                 "{a:1}",
		"{a: 1 b:2}":             "{a:1, b:2}",
		"{a: [1] b:2}":           "{a:[1], b:2}",
		"{a: [1] b:2 c: 1+2}":    "{a:[1], b:2, c:(1+2)}",
		"{a: [1] b:2+n2 c: 1+2}": "{a:[1], b:(2+n2), c:(1+2)}",
		"{a: 1}.a":               "({a:1}.a)",
		`{a: 1}["a"]`:            "({a:1}['a'])",
	}
	for input, want := range tests {
		parser := New(input, testBuiltins())
		parser.advanceTo(0)
		scope := newScope(nil, &Program{})
		scope.set("n1", &Var{Name: "n1", T: NUM_TYPE})
		scope.set("n2", &Var{Name: "n2", T: NUM_TYPE})
		scope.set("s", &Var{Name: "s", T: STRING_TYPE})
		scope.set("b", &Var{Name: "b", T: BOOL_TYPE})
		arrType := &Type{Name: ARRAY, Sub: BOOL_TYPE}
		scope.set("arr", &Var{Name: "arr", T: arrType})
		arrType2 := &Type{Name: ARRAY, Sub: arrType}
		scope.set("arr2", &Var{Name: "arr2", T: arrType2})
		mapType := &Type{Name: MAP, Sub: NUM_TYPE}
		scope.set("map", &Var{Name: "map", T: mapType})
		mapType2 := &Type{Name: MAP, Sub: mapType}
		scope.set("map2", &Var{Name: "map2", T: mapType2})
		arrayMapType := &Type{Name: ARRAY, Sub: mapType}
		scope.set("list", &Var{Name: "list", T: arrayMapType})
		mapArrayType := &Type{Name: MAP, Sub: arrType}
		scope.set("map3", &Var{Name: "map3", T: mapArrayType})

		got := parser.parseTopLevelExpr(scope)
		assertNoParseError(t, parser, input)
		assert.Equal(t, want, got.String())
	}
}

func TestParseTopLevelExpressionErr(t *testing.T) {
	tests := map[string]string{
		"x":        "line 1 column 1: unknown variable name 'x'",
		"+1":       "line 1 column 1: unexpected '+'",
		"* n1":     "line 1 column 1: unexpected '*'",
		"and true": "line 1 column 1: unexpected 'and'",

		"1 +":    "line 1 column 4: unexpected end of input",
		"1 +\n2": "line 1 column 4: unexpected end of line",
		"1 ==":   "line 1 column 5: unexpected end of input",

		"true + false": "line 1 column 6: '+' takes num, string or array type, found bool",
		"true - false": "line 1 column 6: '-' takes num type, found bool",
		"true < false": "line 1 column 6: '<' takes num or string type, found bool",
		"1 and 2":      "line 1 column 3: 'and' takes bool type, found num",
		"1 + false":    "line 1 column 3: mismatched type for +: num, bool",
		"false + 1":    "line 1 column 7: mismatched type for +: bool, num",
		"(1+2":         "line 1 column 5: expected ')', got end of input",
		"(1+2\n)":      "line 1 column 5: expected ')', got end of line",
		"(1+)2":        "line 1 column 4: unexpected ')'",
		"(1+]2":        "line 1 column 4: unexpected ']'",
		"(1+2]":        "line 1 column 5: expected ')', got ']'",

		`"abc"["a"]`:   "line 1 column 11: string index expects num, found string",
		`[1 2 3]["a"]`: "line 1 column 13: array index expects num, found string",
		"{a:2}[2]":     "line 1 column 9: map index expects string, found num",

		"{a:}":    "line 1 column 4: unexpected '}'",
		"{:a}":    "line 1 column 2: expected map key, found ':'",
		"[1 [2]]": "line 1 column 4: only array, string and map type can be indexed, found num",
	}
	for input, wantErr := range tests {
		parser := New(input, testBuiltins())
		parser.advanceTo(0)
		scope := newScope(nil, &Program{})
		scope.set("n1", &Var{Name: "n1", T: NUM_TYPE})

		_ = parser.parseTopLevelExpr(scope)
		assertParseError(t, parser, input)
		assert.Equal(t, wantErr, parser.MaxErrorsString(1), "input: %s\nerrors:\n%s", input, parser.ErrorsString())
	}
}
