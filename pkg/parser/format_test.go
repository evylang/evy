package parser

import (
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestFuncCallStmtFormat(t *testing.T) {
	tests := map[string]string{
		"print 1":                 "print 1\n",
		"print    1  ":            "print 1\n",
		"print  1   // a comment": "print 1 // a comment\n",
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
print 1
else if false
print 2
else
print 3
end
`: `
if true
    print 1
else if false
    print 2
else
    print 3
end
`[1:],
		`if true
  if true
    print 1
  else
    print 1.5
  end
else if false
  if true
    print 2
  end
else
  if true
    print 3
  else if true
    print 4
  end
end
`: `
if true
    if true
        print 1
    else
        print 1.5
    end
else if false
    if true
        print 2
    end
else
    if true
        print 3
    else if true
        print 4
    end
end
`[1:],
		`if true  // if comment
		print 1 // 1 comment
		else if false // else if comment
		print 2 // 2 comment
		else // else comment
		print 3 // 3 comment
		end // end comment
		`: `
if true // if comment
    print 1 // 1 comment
else if false // else if comment
    print 2 // 2 comment
else // else comment
    print 3 // 3 comment
end // end comment
`[1:],
		`if true
print 1
end
`: `
if true
    print 1
end
`[1:],
		`if true  // if comment
		print 1 // 1 comment
		end // end comment
		`: `
if true // if comment
    print 1 // 1 comment
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

func TestFuncDefFormat(t *testing.T) {
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
		"x := a.( num )":      "x := a.(num)",
	}
	for input, want := range tests {
		before := `
n := 1
arr := [1 2 3]
m := {a:1 b:2}
s := "a"
a:any
`[1:]
		after := "\nprint n arr m s a x\n"
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

func TestArrayLiteralFormat(t *testing.T) {
	tests := map[string]string{
		`x := [1 2 3]
		print x
`: `
x := [1 2 3]
print x
`[1:],
		`x := [  1  2    3    ]
		print    x
`: `
x := [1 2 3]
print x
`[1:],
		`x := [
		1
		2
		 ]
		print    x
`: `
x := [
    1
    2
]
print x
`[1:],
		`x := [1
		1.5
		2]
		print    x
`: `
x := [1
    1.5
    2]
print x
`[1:],
		`x := [


		    1


		    2

		  ]
		  print    x
`: `
x := [

    1

    2

]
print x
`[1:],
		`x := [ // comment
// line comment 1
		1 // comment 1

// line comment 2


		2 // comment 2
// line comment 3
		 ]
		print    x
`: `
x := [ // comment
    // line comment 1
    1 // comment 1

    // line comment 2

    2 // comment 2
    // line comment 3
]
print x
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

func TestMapLiteralFormat(t *testing.T) {
	tests := map[string]string{
		"x:={   }":            "x := {}",
		"x:= { a : 1 b:2   }": "x := {a:1 b:2}",
		`
x:= {
			a : 1
			b:2
		}`: `
x := {
    a:1
    b:2
}`,
		`
x:= {			a : 1
			b:2	}`: `
x := {a:1
    b:2}`,
		`
x:{}num
if true
  x= {
			    a : 1
			    b:2
		  }
end`: `
x:{}num
if true
    x = {
        a:1
        b:2
    }
end`,
		`
x:= {			a : 1 // comment 1
			b:2	} // comment 2`: `
x := {a:1 // comment 1
    b:2} // comment 2`,
		`
x:= {


	  a : 1


	  b:2


}`: `
x := {

    a:1

    b:2

}`,
		`
x:= {

    // comment 1
	  a : 1

    // comment 2

	  b:2
	  // comment 3
	  // comment 4
}`: `
x := {

    // comment 1
    a:1

    // comment 2

    b:2
    // comment 3
    // comment 4
}`,
	}

	for input, want := range tests {
		input, want := input+"\nprint x", want+"\nprint x\n"
		t.Run(input, func(t *testing.T) {
			got := testFormat(t, input)
			assert.Equal(t, want, got)
		})
	}
}

func TestArrayMapLiteralFormat(t *testing.T) {
	input := `x := [
  { x:1}
  { x:2}
  { x:3}
  { x:4}
  { x:3}
  { x:2}
  { x:1}
]
print x`
	want := `
x := [
    {x:1}
    {x:2}
    {x:3}
    {x:4}
    {x:3}
    {x:2}
    {x:1}
]
print x
`[1:]
	parser := newParser(input, testBuiltins())
	prog := parser.parse()
	assertNoParseError(t, parser, input)
	got := prog.Format()
	assert.Equal(t, want, got)
}

func TestNLStringLit(t *testing.T) {
	input := `
x := "a\nb"
print x
`[1:]
	want := input
	parser := newParser(input, testBuiltins())
	prog := parser.parse()
	assertNoParseError(t, parser, input)
	got := prog.Format()
	assert.Equal(t, want, got)
}

func TestUnaryOP(t *testing.T) {
	input := `
func check:bool
    return false
end

while !(check)
    print "x"
end
`[1:]
	want := input
	parser := newParser(input, testBuiltins())
	prog := parser.parse()
	assertNoParseError(t, parser, input)
	got := prog.Format()
	assert.Equal(t, want, got)
}

func TestNLInsertion(t *testing.T) {
	tests := map[string]string{
		`func f1
	print 1
end // needs nl after this line, stmt IDX 0
func f2
    print 2
end`: `
func f1
    print 1
end // needs nl after this line, stmt IDX 0

func f2
    print 2
end
`[1:],
		`
func f1
    print 1
end // needs nl after this line, stmt IDX 0
func f2
    print 2
end`: `
func f1
    print 1
end // needs nl after this line, stmt IDX 0

func f2
    print 2
end
`,
		`
func f1
    print 1
end // needs nl after this line, stmt IDX 0
// f2 comment
func f2
    print 2
end
`: `
func f1
    print 1
end // needs nl after this line, stmt IDX 0

// f2 comment
func f2
    print 2
end
`,
		`
func f1
    print 1
end // needs nl after this line, stmt IDX 0
// f2 comment
// f2 comment continued
on down
    print 2
end`: `
func f1
    print 1
end // needs nl after this line, stmt IDX 0

// f2 comment
// f2 comment continued
on down
    print 2
end
`,
		`
// f1 comment
func f1
    print 1
end // needs nl after this line, stmt IDX 0
print 1
// f2 comment
// f2 comment continued
on down
    print 2
end`: `
// f1 comment
func f1
    print 1
end // needs nl after this line, stmt IDX 0

print 1

// f2 comment
// f2 comment continued
on down
    print 2
end
`,
		`
a := 1
b := 2
func fn
    print "fn" a b
end
func fn2
    print "fn2" a b
end
fn
`: `
a := 1
b := 2

func fn
    print "fn" a b
end

func fn2
    print "fn2" a b
end

fn
`, `
a := 1
func fn
    print "fn" a
end
`: `
a := 1

func fn
    print "fn" a
end
`, `
a := 1
b := 2
func fn
    print "fn" a b
end
`: `
a := 1
b := 2

func fn
    print "fn" a b
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
	parser := newParser(input, testBuiltins())
	prog := parser.parse()
	assertNoParseError(t, parser, input)
	return prog.Format()
}
