package parser

import (
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestReturnStmtFormat(t *testing.T) {
	tests := map[string]string{
		"return 1":                 "return 1\n",
		"return    1  ":            "return 1\n",
		"return  1   // a comment": "return 1 // a comment\n",
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestWhileStmtFormat(t *testing.T) {
	tests := map[string]string{
		`while true
break
end`: `
while true
    break
end
`[1:],

		`while true  // while comment
// line comment
break     // break comment
end// end comment`: `
while true // while comment
    // line comment
    break // break comment
end // end comment
`[1:],
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestIfStmtFormat(t *testing.T) {
	tests := map[string]string{
		`if true
return 1
else if false
return 2
else
return 3
end
`: `
if true
    return 1
else if false
    return 2
else
    return 3
end
`[1:],
		`if true
  if true
    return 1
  else
    return 1.5
  end
else if false
  if true
    return 2
  end
else
  if true
    return 3
  else if true
    return 4
  end
end
`: `
if true
    if true
        return 1
    else
        return 1.5
    end
else if false
    if true
        return 2
    end
else
    if true
        return 3
    else if true
        return 4
    end
end
`[1:],
		`if true  // if comment
		return 1 // 1 comment
		else if false // else if comment
		return 2 // 2 comment
		else // else comment
		return 3 // 3 comment
		end // end comment
		`: `
if true // if comment
    return 1 // 1 comment
else if false // else if comment
    return 2 // 2 comment
else // else comment
    return 3 // 3 comment
end // end comment
`[1:],
		`if true
return 1
end
`: `
if true
    return 1
end
`[1:],
		`if true  // if comment
		return 1 // 1 comment
		end // end comment
		`: `
if true // if comment
    return 1 // 1 comment
end // end comment
`[1:],
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestForStmtFormat(t *testing.T) {
	tests := map[string]string{
		`for   i:=   range 10
print    i
end
`: `
for i := range 10
    print i
end
`[1:],
		`for   i:=   range 10   20
print    i
end
`: `
for i := range 10 20
    print i
end
`[1:],
		`for   i:=   range 10   20  3
print    i
end
`: `
for i := range 10 20 3
    print i
end
`[1:],
		`for   i:=   range 10   20  3 // for comment
print    i  // print comment
end // end comment
`: `
for i := range 10 20 3 // for comment
    print i // print comment
end // end comment
`[1:],
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestDeclAssignStmtFormat(t *testing.T) {
	tests := map[string]string{
		`i   :=  7
		i  =   5
		print   i
`: `
i := 7
i = 5
print i
`[1:],
		`i :  num
		print   i
`: `
i:num
print i
`[1:],
		`i   :=  7   // comment i
		print   i
`: `
i := 7 // comment i
print i
`[1:],
		`i :  num // comment decl
i=5 // comment assign
		print   i
`: `
i:num // comment decl
i = 5 // comment assign
print i
`[1:],
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestFuncDeclFormat(t *testing.T) {
	tests := map[string]string{
		`func fox     // func
		print   "🦊"  // print
		end  // end
`: `
func fox // func
    print "🦊" // print
end // end
`[1:],
		`func fox : string a:num b : bool     // func
		print   a   b // print
		return  "🦊"   // return
		end  // end
`: `
func fox:string a:num b:bool // func
    print a b // print
    return "🦊" // return
end // end
`[1:],
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestEventHandlerFormat(t *testing.T) {
	tests := map[string]string{
		`on down // on
		print   "🦊"  // print
		end  // end
`: `
on down // on
    print "🦊" // print
end // end
`[1:],
		`on down  x:num y:num   // on
		print  x y // print
		end  // end
`: `
on down x:num y:num // on
    print x y // print
end // end
`[1:],
	}

	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestPrintExpressionFormat(t *testing.T) {
	tests := map[string]string{
		"print  1+2":                 "print 1+2\n",
		"print  1+2 // comment":      "print 1+2 // comment\n",
		"print  (true  or  false)  ": "print (true or false)\n",
		"print  1+2*3":               "print 1+2*3\n",
		"print   [1 2 3][0]":         "print [1 2 3][0]\n",
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestExpressionFormat(t *testing.T) {
	tests := map[string]string{
		"x := 1+2":            "x := 1 + 2",
		"x := 1+2*3/n":        "x := 1 + 2 * 3 / n",
		"x := [[1] [2]] ":     "x := [[1] [2]]",
		"x := arr[n + 2] ":    "x := arr[n + 2]",
		"x := [ 2+n   3*n  ]": "x := [2+n 3*n]",
		"x := s+s":            "x := s + s",
		"x := s[  0  ]":       "x := s[0]",
		"x := m.a + n":        "x := m.a + n",
		"x := [m.a+n]":        "x := [m.a+n]",
		"x := s[m.a+n]":       "x := s[m.a + n]",
		"x := arr[1:n+3]":     "x := arr[1:n + 3]",
		"x := arr[ :n+3]":     "x := arr[:n + 3]",
		"x := arr[ n+3: ]":    "x := arr[n + 3:]",
	}
	for input, want := range tests {
		before := `
n := 1
arr := [1 2 3]
m := {a:1 b:2}
s := "a"
`[1:]
		after := "\nprint n arr m s x\n"
		input := before + input + after
		want := before + want + after
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestTestEmptyStmtFormat(t *testing.T) {
	tests := map[string]string{
		"":       "\n",
		"\n":     "\n",
		"\n\n":   "\n",
		"\n\n\n": "\n",

		"//asdf":             "//asdf\n",
		"//asdf\n\n":         "//asdf\n\n",
		"//asdf\n\n\n":       "//asdf\n\n",
		"\n//asdf\n\n\n":     "\n//asdf\n\n",
		"\n\n//asdf\n\n\n":   "\n//asdf\n\n",
		"\n\n\n//asdf\n\n\n": "\n//asdf\n\n",

		`
if true


  // tests
// test1

  print 1
end`: `
if true

    // tests
    // test1

    print 1
end
`,
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func testFormat(t *testing.T, input string) string {
	t.Helper()
	parser := New(input, testBuiltins())
	prog := parser.Parse()
	assertNoParseError(t, parser, input)
	return prog.Format()
}
