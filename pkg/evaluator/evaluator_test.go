package evaluator

import (
	"bytes"
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestBasicEval(t *testing.T) {
	in := "a:=1\n print a 2"
	want := "1 2\n"
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(in, fn)
	assert.Equal(t, want, b.String())
}

func TestParseDeclaration(t *testing.T) {
	tests := map[string]string{
		"a:=1":          "1",
		`a:="abc"`:      "abc",
		`a:=true`:       "true",
		`a:= len "abc"`: "3",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print a"
			b := bytes.Buffer{}
			fn := func(s string) { b.WriteString(s) }
			Run(in, fn)
			assert.Equal(t, want+"\n", b.String())
		})
	}
}

func TestReturn(t *testing.T) {
	prog := `
func fox:string
    return "🦊"
end

func fox2
    if true
        print "🦊2"
        return
    end
    print "💣"
end

f := fox
print f
print f f
fox2
`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(prog, fn)
	want := "🦊\n🦊 🦊\n🦊2\n"
	assert.Equal(t, want, b.String())
}

func TestReturnScope(t *testing.T) {
	prog := `
f := 1

func fox1:string
    f := "🦊"
    return f
end

func fox2:string
    return fox1
end

print f
f1 := fox1
print f1
f2 := fox2
print f2
`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(prog, fn)
	want := "1\n🦊\n🦊\n"
	assert.Equal(t, want, b.String())
}

func TestBreak(t *testing.T) {
	tests := []string{
		`
while true
    print "🎈"
    break
end
`, `
while true
    print "🎈"
    if true
        break
    end
    print "💣"
end
`, `
stop := false
while true
    if stop
        print "🎈"
        break
    end
    stop = true
end
`, `
continue := true
while true
    if continue
        print "🎈"
    else
        break
    end
    continue = false
end
`,
	}
	for _, input := range tests {
		b := bytes.Buffer{}
		fn := func(s string) { b.WriteString(s) }
		Run(input, fn)
		want := "🎈\n"
		assert.Equal(t, want, b.String(), input)
	}
}

func TestAssignment(t *testing.T) {
	prog := `
f1:num
f2:num
f3 := 3
print f1 f2 f3
f1 = 1
print f1 f2 f3
f1 = f3
f2 = f1
f3 = 4
print f1 f2 f3
`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(prog, fn)
	want := "0 0 3\n1 0 3\n3 3 4\n"
	assert.Equal(t, want, b.String())
}

func TestAssignmentAny(t *testing.T) {
	prog := `
f1:any
f2:num
print f1 f2
f1 = f2
print f1 f2
f1 = fox
print f1 f2

func fox:string
    return "🦊"
end

`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(prog, fn)
	want := "false 0\n0 0\n🦊 0\n"
	assert.Equal(t, want, b.String())
}

func TestIf(t *testing.T) {
	tests := []string{
		`
if true
    print "🎈"
else
    print "💣"
end
`,
		`
x := "💣"
if true
    x = "🎈"
end
print x
`,
		`
if false
    print "💣"
else
    if true
        print "🎈"
    end
end
`,
		`
if true
    if false
        print "💣1"
    else if true
        print "🎈"
    else if true
        print "💣2"
    else
        print "💣3"
    end
else
    print "💣4"
end
`,
	}
	for _, input := range tests {
		b := bytes.Buffer{}
		fn := func(s string) { b.WriteString(s) }
		Run(input, fn)
		assert.Equal(t, "🎈\n", b.String(), "input: %s", input)
	}
}

func TestWhile(t *testing.T) {
	input := `
x := true
while x
	print "🍭"
	x = false
end

one_more := true
two_more := true
func has_more:bool
	if one_more
		if two_more
			two_more = false
			return true
		else
			one_more = false
			return true
		end
	end
	return false
end

one_more = true
while has_more
	print "🎈"
end

while has_more
	print "💣"
end
`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(input, fn)
	assert.Equal(t, "🍭\n🎈\n🎈\n", b.String())
}

func TestExpr(t *testing.T) {
	tests := map[string]string{
		"a := 1 + 2 * 2":                    "5",
		"a := (1 + 2) * 2":                  "6",
		"a := (1 + 2) / 2":                  "1.5",
		"a := (1 + 2) / 2 > 1":              "true",
		"a := (1 + 2) / 2 > 1 and 2 == 2*2": "false",
		"a := (1 + 2) / 2 < 1 or 2 == 2*2":  "false",
		"a := (1 + 2) / 2 < 1 or 2 != 2*2":  "true",
		`a := "abc" + "d"`:                  "abcd",
		`a := "abc" + "d" < "efg"`:          "true",
		`a := "abc" + "d" == "abcd"`:        "true",
		`a := "abc" + "d" != "abcd"`:        "false",
		`a := !(1 == 1)`:                    "false",
		`a := -(3 + 5)`:                     "-8",
		`a := -3 +5`:                        "2",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print a"
			b := bytes.Buffer{}
			fn := func(s string) { b.WriteString(s) }
			Run(in, fn)
			assert.Equal(t, want+"\n", b.String())
		})
	}
}

func TestArrayLit(t *testing.T) {
	tests := map[string]string{
		"a := [1]":     "[1]",
		"a := []":      "[]",
		"a := [1 2]":   "[1 2]",
		"a := [1 1+1]": "[1 2]",
		`
b := 3
a := [1 1+1 b]`: "[1 2 3]",
		`
func three:num
    return 3
end
a := [1 1+1 (three)]`: "[1 2 3]",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print a"
			b := bytes.Buffer{}
			fn := func(s string) { b.WriteString(s) }
			Run(in, fn)
			assert.Equal(t, want+"\n", b.String())
		})
	}
}

func TestMapLit(t *testing.T) {
	tests := map[string]string{
		"a := {n:1}":                 "{n:1}",
		"a := {}":                    "{}",
		`a := {name:"fox" age:42}`:   "{name:fox age:42}",
		`a := {name:"fox" age:40+2}`: "{name:fox age:42}",
		`
b := 2
a := {name:"fox" age:40+b}`: "{name:fox age:42}",
		`
func three:num
    return 3
end
a := {name:"fox" age:39+(three)}`: "{name:fox age:42}",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print a"
			b := bytes.Buffer{}
			fn := func(s string) { b.WriteString(s) }
			Run(in, fn)
			assert.Equal(t, want+"\n", b.String())
		})
	}
}

func TestDemo(t *testing.T) {
	prog := `
move 10 10
line 20 20

x := 12
print "x:" x
if x > 10
    print "🍦 big x"
end`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(prog, fn)
	want := `
'move' not yet implemented
'line' not yet implemented
x: 12
🍦 big x
`[1:]
	assert.Equal(t, want, b.String())
}
