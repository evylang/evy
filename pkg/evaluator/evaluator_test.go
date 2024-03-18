package evaluator

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

type testRT struct {
	UnimplementedRuntime
	b bytes.Buffer
}

func (rt *testRT) Print(s string) {
	rt.b.WriteString(s)
}

func (*testRT) Yielder() Yielder {
	return nil
}

func run(input string) string {
	rt := &testRT{}
	rt.UnimplementedRuntime.print = rt.Print
	eval := NewEvaluator(rt)
	err := eval.Run(input)
	if err != nil {
		return err.Error()
	}
	return rt.b.String()
}

func TestBasicEval(t *testing.T) {
	in := "a:=1\n print a 2"
	want := "1 2\n"
	got := run(in)
	assert.Equal(t, want, got)
}

func TestParseDecl(t *testing.T) {
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
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestFuncScope(t *testing.T) {
	prog := `
f := "🦊"

func outer
    f := "🐤"
    print f
    if true
        f := "🎈"
        print f
        inner
        print f
    end
    print f
end

func inner
    print f
end

print f
outer
print f
`
	want := "🦊\n🐤\n🎈\n🦊\n🎈\n🐤\n🦊\n"
	got := run(prog)
	assert.Equal(t, want, got)
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
	want := "🦊\n🦊 🦊\n🦊2\n"
	got := run(prog)
	assert.Equal(t, want, got)
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
	want := "1\n🦊\n🦊\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestErrVarNotSet(t *testing.T) {
	prog := `
fn "hello"

v := "world"

func fn s:string
    print s v  // line 7
end
`

	rt := &testRT{}
	rt.UnimplementedRuntime.print = rt.Print
	eval := NewEvaluator(&testRT{})
	err := eval.Run(prog)
	assert.Equal(t, true, errors.Is(err, ErrVarNotSet))
	evalErr := &Error{}
	assert.Equal(t, true, errors.As(err, &evalErr))
	assert.Equal(t, "line 7 column 13", evalErr.Token.Location())
}

func TestLenString(t *testing.T) {
	prog := `
print (len "🌟✨🌙🪐") // 4 runes, 4 graphemes
print (len "⭐️") // 2 runes(!), 1 cluster/grapheme`
	want := "4\n2\n"
	got := run(prog)
	assert.Equal(t, want, got)
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
`,
		`
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
`, `
while true
    while true
        break
    end
    print "🎈"
    break
end
`, `
for range 3
    for range 3
        break
    end
    print "🎈"
    break
end
`,
	}
	for _, input := range tests {
		want := "🎈\n"
		got := run(input)
		assert.Equal(t, want, got, input)
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
	want := "0 0 3\n1 0 3\n3 3 4\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAssignmentAny(t *testing.T) {
	prog := `
func fox:string
    return "🦊"
end

func lol_any:any
    return "🍭"
end

f1:any
f2:num
print "1" f1 f2

f1 = f2
print "2" f1 f2

f1 = fox
print "3" f1 f2

f1 = lol_any
print "4" f1

f3 := f1
print "5" f3==f1

f4:any
f4 = f1
print "6" f4==f1
`
	wants := []string{
		"1 false 0",
		"2 0 0",
		"3 🦊 0",
		"4 🍭",
		"5 true",
		"6 true",
		"",
	}
	want := strings.Join(wants, "\n")
	got := run(prog)

	assert.Equal(t, want, got)
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
		got := run(input)
		assert.Equal(t, "🎈\n", got, "input: %s", input)
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
	got := run(input)
	assert.Equal(t, "🍭\n🎈\n🎈\n", got)
}

func TestExpr(t *testing.T) {
	tests := map[string]string{
		"a := 1 + 2 * 2":                    "5",
		"a := 6 % 4":                        "2",
		"a := 6.3 % 4.1":                    "2.2",
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
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestShortCircuit(t *testing.T) {
	tests := map[string]string{
		"a := false and (side_effect false)":           "false",
		"a := false and (side_effect false) and false": "false",
		"a := false and false and (side_effect false)": "false",
		"a := true or (side_effect true) and true":     "true",
		"a := true or false or (side_effect true)":     "true",
		"a := true or (side_effect true)":              "true",
		"a := true and (side_effect true)":             "side-effect\ntrue",
		"a := false or (side_effect false)":            "side-effect\nfalse",
		"a := true and (side_effect false)":            "side-effect\nfalse",
		"a := false or (side_effect true)":             "side-effect\ntrue",
	}
	const preamble = `
func side_effect:bool v:bool
	print "side-effect"
	return v
end`
	for in, want := range tests {
		in, want := preamble+"\n"+in+"\n print a", want
		t.Run(in, func(t *testing.T) {
			got := run(in)
			assert.Equal(t, want+"\n", got)
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
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestIndex(t *testing.T) {
	tests := map[string]string{
		// x := ["a","b","c"]; x = "abc"
		"print x[0]":  "a",
		"print x[1]":  "b",
		"print x[2]":  "c",
		"print x[-1]": "c",
		"print x[-2]": "b",
		"print x[-3]": "a",
		`
		n1 := 1
		print x[n1 - 1] x[1 + n1]
		`: "a c",
		`
		m := {a: "bingo"}
		print m[x[0]]
		`: "bingo",
	}
	for in, want := range tests {
		in, want := in, want
		for _, decl := range []string{`x := ["a" "b" "c"]`, `x := "abc"`} {
			input := decl + "\n" + in
			t.Run(input, func(t *testing.T) {
				got := run(input)
				assert.Equal(t, want+"\n", got)
			})
		}
	}
}

func TestDoubleIndex(t *testing.T) {
	tests := map[string]string{
		`
		x := [ [1 2 3] ([4 5 6]) ]
		x[0][1] = 99
		print x
		`: "[[1 99 3] [4 5 6]]",
	}
	for in, want := range tests {
		t.Run(in, func(t *testing.T) {
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestIndexErr(t *testing.T) {
	tests := map[string]string{
		// x := ["a","b","c"]; x = "abc"
		"print x[3]":  "line 2 column 8: panic: index out of bounds: 3",
		"print x[-4]": "line 2 column 8: panic: index out of bounds: -4",
		`m := {}
		print m[x[1]]`: `line 3 column 10: panic: no value for map key: "b"`,
	}
	for in, want := range tests {
		in, want := in, want
		for _, decl := range []string{`x := ["a" "b" "c"]`, `x := "abc"`} {
			input := decl + "\n" + in
			t.Run(input, func(t *testing.T) {
				got := run(input)
				assert.Equal(t, want, got)
			})
		}
	}
}

func TestMapLit(t *testing.T) {
	tests := map[string]string{
		"a := {n:1}":                 "{n:1}",
		"a := {for:1}":               "{for:1}",
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
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestDot(t *testing.T) {
	tests := map[string]string{
		// m := {name: "Greta"}
		"print m.name":    "Greta",
		"print m.for":     "FFF",
		`print m["name"]`: "Greta",
		`print m["for"]`:  "FFF",
		`s := "name"
		print m[s]`: "Greta",
		`s := "for"
		print m[s]`: "FFF",
	}
	for in, want := range tests {
		in, want := in, want
		input := `m := {name: "Greta" for: "FFF"}` + "\n" + in
		t.Run(input, func(t *testing.T) {
			got := run(input)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestDotErr(t *testing.T) {
	in := `
m := {a:1}
print m.missing_index
`
	want := `line 3 column 8: panic: no value for map key: "missing_index"`
	got := run(in)
	assert.Equal(t, want, got)
}

func TestArrayConcatenation(t *testing.T) {
	prog := `
arr1 := [1]
arr2 := arr1
arr3 := arr1 + arr1
arr4 := arr1 + [2]
arr5 := arr1 + []
arr6 := [] + []
print "1 arr1" arr1
print "1 arr2" arr2
print "1 arr3" arr3
print "1 arr4" arr4
print "1 arr5" arr5
print "1 arr6" arr6
print

arr1[0] = 2
print "2 arr1" arr1
print "2 arr2" arr2
print "2 arr3" arr3
print "2 arr4" arr4
print "2 arr5" arr5
`
	out := run(prog)
	want := []string{
		"1 arr1 [1]",
		"1 arr2 [1]",
		"1 arr3 [1 1]",
		"1 arr4 [1 2]",
		"1 arr5 [1]",
		"1 arr6 []",
		"",
		"2 arr1 [2]",
		"2 arr2 [2]",
		"2 arr3 [1 1]",
		"2 arr4 [1 2]",
		"2 arr5 [1]",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestArraySlice(t *testing.T) {
	prog := `
arr := [1 2 3]
print "1" arr[1:3]
print "2" arr[1:]
print "3" arr[1:2]
print "4" arr[1:1]
print "5" arr[:1]
print "6" arr[3:]
print "7" arr[3:3]
print

arr2 := arr[:]
arr2[0] = 11
print "8" arr arr2
`
	out := run(prog)
	want := []string{
		"1 [2 3]",
		"2 [2 3]",
		"3 [2]",
		"4 []",
		"5 [1]",
		"6 []",
		"7 []",
		"",
		"8 [1 2 3] [11 2 3]",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestStringSlice(t *testing.T) {
	prog := `
s := "abc"
print "1" s[1:3]
print "2" s[1:]
print "3" s[1:2]
print "4" s[1:1]
print "5" s[:1]
print "6" s[3:]
print "7" s[3:3]
print

s2 := "A" + s[1:]
print "8" s s2
`
	out := run(prog)
	want := []string{
		"1 bc",
		"2 bc",
		"3 b",
		"4 ",
		"5 a",
		"6 ",
		"7 ",
		"",
		"8 abc Abc",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestForStepRange(t *testing.T) {
	prog := `
for i := range 2
	print "🎈" i
end
for range 2
	print "🦈"
end
for i := range -1 1
	print "🐣" i
end
for i := range 2 6 2
	print "🍭" i
end
for i := range 5 3 -1
	print "🦊" i
end
for i := range 3 5 -1
	print "1💣" i
end
for i := range 3 -1 1
	print "2💣" i
end
for i := range 3 -1
	print "3💣" i
end
`
	out := run(prog)
	want := []string{
		"🎈 0",
		"🎈 1",
		"🦈",
		"🦈",
		"🐣 -1",
		"🐣 0",
		"🍭 2",
		"🍭 4",
		"🦊 5",
		"🦊 4",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestForArray(t *testing.T) {
	prog := `
for x := range [0 1]
	print "🎈" x
end
for range [0 1]
	print "🦊"
end
for i := range []
	print "💣" i
end
`
	out := run(prog)
	want := []string{
		"🎈 0",
		"🎈 1",
		"🦊",
		"🦊",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestForString(t *testing.T) {
	prog := `
for x := range "abc"
	print "🎈" x
end
for range "ab"
	print "🦊"
end
for i := range ""
	print "💣" i
end
`
	out := run(prog)
	want := []string{
		"🎈 a",
		"🎈 b",
		"🎈 c",
		"🦊",
		"🦊",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestForMap(t *testing.T) {
	prog := `
m := {a:1 b:2}
for x := range m
	print "🎈" x  m[x]
end
for range m
	print "🦊"
end
for i := range {}
	print "💣" i
end
`
	out := run(prog)
	want := []string{
		"🎈 a 1",
		"🎈 b 2",
		"🦊",
		"🦊",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestMap(t *testing.T) {
	prog := `
m1 := {a:1 b:2}
m2 := m1
print "1" m1 m2

m2.a = 10
m1["b"] = 20
print "2" m1 m2

m2.c = 3
m1["d"] = 4
print "3" m1 m2

m4 := {}
m4.a = 1
m4["b"] = 2
print "4" m4

m5 := {}
m5.a = 1
m5.b = {c:99}
//m5.b.c = 2 // parse error: need to cast any to map...
print "5" m5

m6:{}{}num
m6.a = {A :1}
m6.b = {}
m6.b.c = 2
print "6" m6
`
	out := run(prog)
	want := []string{
		"1 {a:1 b:2} {a:1 b:2}",
		"2 {a:10 b:20} {a:10 b:20}",
		"3 {a:10 b:20 c:3 d:4} {a:10 b:20 c:3 d:4}",
		"4 {a:1 b:2}",
		"5 {a:1 b:{c:99}}",
		"6 {a:{A:1} b:{c:2}}",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestMapErr(t *testing.T) {
	in := `
m := {}
m.a = 1
m.b.c = 2
`
	got := run(in)
	want := `line 4 column 4: field access with "." expects map type, found any`
	assert.Equal(t, want, got)
}

func TestHas(t *testing.T) {
	prog := `
m := {a:1 b:2}
print (has m "a")
print (has m "MISSING")
`
	out := run(prog)
	want := []string{
		"true",
		"false",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestHasErr(t *testing.T) {
	prog := `
has ["a"] "a" // cannot run "has" on array
`
	want := `line 2 column 5: "has" takes 1st argument of type {}, found []string`
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestDel(t *testing.T) {
	prog := `
m1 := {a:1 b:2}
m2 := m1
print "1" m1 m2
del m1 "a"
print "2" m1 m2
del m1 "MISSING"
print "3" m1 m2
del m2 "b"
print "4" m1 m2
`
	out := run(prog)
	want := []string{
		"1 {a:1 b:2} {a:1 b:2}",
		"2 {b:2} {b:2}",
		"3 {b:2} {b:2}",
		"4 {} {}",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestDelErr(t *testing.T) {
	prog := `
del ["a"] "a" // cannot delete from array
`
	want := `line 2 column 5: "del" takes 1st argument of type {}, found []string`
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestJoin(t *testing.T) {
	prog := `
print (join [1 true "x"] ", ")
`
	want := "1, true, x\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestSprint(t *testing.T) {
	prog := `
s := sprint 1 [2] "x"
print (s)
`
	want := "1 [2] x\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestSprintf(t *testing.T) {
	prog := `
print "1:" (sprintf "-%4.1f-" 1)
print "2:" (sprintf "%v %v %v %v %v" 42 true "🐥" [1 "b"] {name: "🦊"})
`
	out := run(prog)
	want := []string{
		"1: - 1.0-",
		"2: 42 true 🐥 [1 b] {name:🦊}",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestPrintf(t *testing.T) {
	prog := `
printf "1: -%4.1f-\n" 1
printf "2: %v %v %v %v %v\n" 42 true "🐥" [1 "b"] {name: "🦊"}
`
	out := run(prog)

	want := []string{
		"1: - 1.0-",
		"2: 42 true 🐥 [1 b] {name:🦊}",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestParamAssign(t *testing.T) {
	prog := `
x := 1
f x
x = x + 1
f x

func f n:num
	n = n*10
	print n x
end`

	want := "10 1\n20 2\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAssign2(t *testing.T) {
	prog := `
x := 1
n := x
n = n * 10
print x
`
	want := "1\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAssign3(t *testing.T) {
	prog := `
x:num
x = 1
n:num
n = x
n = n * 10
print x
`
	want := "1\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestSplit(t *testing.T) {
	prog := `
print (split "a, b, c" ", ")
`
	want := "[a b c]\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyAssignment(t *testing.T) {
	prog := `
a := 1
b:any
b = a
a = 2
print a b
`
	want := "2 1\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestCompositeAssignment(t *testing.T) {
	prog := `
n := 1
a := [n n]
m := {n: n}
n = 2
print n a m`
	want := "2 [1 1] {n:1}\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestTypeAssertion(t *testing.T) {
	prog := `
a := {v:[1 2 3]}
b:any
b = a
c := b.({}[]num)
print c`
	want := "{v:[1 2 3]}\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestTypeAssertionErr(t *testing.T) {
	// Note the zero value for any is a false bool
	prog := `
a:any
b := a.(num)
print b`
	want := "line 3 column 7: panic: error converting any to type: expected num, found bool"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestTypeof(t *testing.T) {
	tests := map[string]string{
		"a:bool":                   "bool",
		"a:num":                    "num",
		"a:string":                 "string",
		"a:any":                    "bool", // zero value for any is bool false
		"a:[]num":                  "[]num",
		"a:{}string":               "{}string",
		"a:[]{}[]{}bool":           "[]{}[]{}bool",
		"a := 1":                   "num",
		"a := true":                "bool",
		`a := "hello"`:             "string",
		"a := [1 2 3]":             "[]num",
		"a := [{a:1}]":             "[]{}num",
		"a := [{a:1} {b:true}]":    "[]{}any",
		"a := [{a:1} {b:true} {}]": "[]{}any",
		"a := [{a:1} {b:2} {}]":    "[]{}num",
		"a := [{a:1} {b:2} false]": "[]any",
		"a := {a:1 b:[2]}":         "{}any",
		"a := {a:[1 2 3] b:[]}":    "{}[]num",
		"a := []":                  "[]any",
		"a := {}":                  "{}any",
		"a:[]num \n a = []":        "[]num",
		"a:{}num \n a = {}":        "{}num",
		"a := [] + [true]":         "[]bool",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print (typeof a)"
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}

	tests = map[string]string{
		`a := [ "abc" 1 ]`: "string",
		`a := [ 1 "abc" ]`: "num",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print (typeof a[0])"
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}

	tests = map[string]string{
		`print (typeof [])`:        "[]any",
		`print (typeof [[]])`:      "[][]any",
		`print (typeof {})`:        "{}any",
		`print (typeof []+[true])`: "[]bool",
		`print (typeof [true]+[])`: "[]bool",
		`print (typeof []+[])`:     "[]any",
		`print (typeof ([]))`:      "[]any",
		`print (typeof [2]+[])`:    "[]num",
		`print (typeof []+[2])`:    "[]num",
		`print (typeof []+[[2]])`:  "[][]num",
		`print (typeof [[2]]+[])`:  "[][]num",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestTypeofParam(t *testing.T) {
	tests := map[string]string{
		`
func fn x:[]num
    print (typeof x)
end
fn []
`: "[]num",
		`
func fn x:{}num
    print (typeof x)
end
fn {}
`: "{}num",
		`
func fn x:[][]num
    print (typeof x)
end
fn []
`: "[][]num",
		`
func fn x:{}{}num
    print (typeof x)
end
fn {}
`: "{}{}num",
		`
func fn x:[]{}num
    print (typeof x)
end
fn []
`: "[]{}num",
		`
func fn x:{}[]num
    print (typeof x)
end
fn {}
`: "{}[]num",
		`
func fn x:num...
    print (typeof x)
end
fn
`: "[]num",
		`
func fn x:[]num...
    print (typeof x)
end
fn
fn []
`: "[][]num\n[][]num",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestStr2BoolNum(t *testing.T) {
	prog := `
print 1 (str2bool "1") err
print 2 (str2bool "TRUE") err
print 3 (str2bool "true") err
print 4 (str2bool "False") err
print 5 (str2num "1") err
print 6 (str2num "-1.7") err
`
	out := run(prog)
	want := []string{
		"1 true false",
		"2 true false",
		"3 true false",
		"4 false false",
		"5 1 false",
		"6 -1.7 false",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestStr2BoolNumErr(t *testing.T) {
	prog := `
print 1 (str2bool "BAD") err
print 2 errmsg
str2bool "true"
print 3 err  // check reset
print 4 (str2num "BAD") err
print 5 errmsg
`
	out := run(prog)
	want := []string{
		"1 false true",
		`2 str2bool: cannot parse "BAD"`,
		"3 false",
		"4 0 true",
		`5 str2num: cannot parse "BAD"`,
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestVariadic(t *testing.T) {
	prog := `
func fox nums:num...
  for n := range nums
    print n "🦊"
  end
end

func fox2 strings:string...
  for s := range strings
    print s "🦊"
  end
end

func fox3 anys:any...
  for a := range anys
    print a "🦊"
  end
end

fox 1 2 3
fox2 "a" "b"
fox3 [1 2] true
`
	out := run(prog)
	want := []string{
		"1 🦊",
		"2 🦊",
		"3 🦊",
		"a 🦊",
		"b 🦊",
		"[1 2] 🦊",
		"true 🦊",
		"",
	}
	got := strings.Split(out, "\n")
	assert.Equal(t, len(want), len(got), out)
	for i := range want {
		assert.Equal(t, want[i], got[i])
	}
}

func TestIndexing(t *testing.T) {
	tests := map[string]string{
		"print [1 2] [1]":          "[1 2] [1]",
		"print [1 2][1]":           "2",
		"print {} []":              "{} []",
		"print [] []":              "[] []",
		"print [] {}":              "[] {}",
		"print {} {}":              "{} {}",
		`print {a:1}["a"]`:         "1",
		`print {a:1} ["a"]`:        "{a:1} [a]",
		`print ( sin ( sin 0 ) ) `: "0",
		`
func fn:{}num
	return {a:1}
end
print (fn)["a"]
print (fn) ["a"]
`: "1\n{a:1} [a]",
		`
func fn:string
	return "abc"
end
print (fn)[1]
print (fn) [1]
`: "b\nabc [1]",
		`a:any
a = [1 2]
print a.([]num)[1]
print a.([]num) [1]
`: "2\n[1 2] [1]",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			got := run(in)
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestAnyArray(t *testing.T) {
	prog := `
x:[]any
x = [1 2 true]
print x
x[1] = [3 4 5]
print x
`
	want := "[1 2 true]\n[1 [3 4 5] true]\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyMap(t *testing.T) {
	prog := `
x:{}any
x = {a:1 b:2 c:true}
print x
x["a"] = [3 4 5]
x.b = [7 8]
print x
`
	want := "{a:1 b:2 c:true}\n{a:[3 4 5] b:[7 8] c:true}\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyBoolArray(t *testing.T) {
	prog := `
x:[]any
x = [true false]
print x
x[1] = [3 4 5]
print x
`
	want := "[true false]\n[true [3 4 5]]\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestNestedAnyArray(t *testing.T) {
	prog := `
x:[][]any
a:any
a = 1
x = [[a 2 true]]
print x
x = [[a 2]]
print x
a = 5
x[0][1] = [3 4 a]
print x
`
	want := "[[1 2 true]]\n[[1 2]]\n[[1 [3 4 5]]]\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyMapTypeOf(t *testing.T) {
	prog := `
x:{}any
x = {a:true}
print (typeof x) x
x = {a:1 b:2 c:true}
print x
x.b = [3 4 5]
print x
`
	want := "{}any {a:true}\n{a:1 b:2 c:true}\n{a:1 b:[3 4 5] c:true}\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestNestedAnyMap(t *testing.T) {
	prog := `
x:{}{}any
a:any
a = 1
x = {A:{A:true}}
print (typeof x) x
x = {A:{A:a B:2 C:true}}
print x
a = 5
x.A.B = [3 4 a]
print x
`
	want := "{}{}any {A:{A:true}}\n{A:{A:1 B:2 C:true}}\n{A:{A:1 B:[3 4 5] C:true}}\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyParamType(t *testing.T) {
	prog := `
func f a:any
    if (typeof a) == "num"
        n := a.(num)
        print "num" n
    else if (typeof a) == "[]num"
        na := a.([]num)
        print "[]num" na
    else if (typeof a) == "[]any"
        aa := a.([]any)
        print "[]any" aa
    else if (typeof a) == "[][]any"
        aa := a.([][]any)
        print "[][]any" aa
    else
        print "typeof:" (typeof a)
    end
end
f 1
f [1 2]
f []
f [[] {}]
f [[]]
f [{}]
f [[1]]
a:any
f a
`
	want := `
num 1
[]num [1 2]
[]any []
[]any [[] {}]
[][]any [[]]
typeof: []{}any
typeof: [][]num
typeof: bool
`[1:]
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyVariadicParamType(t *testing.T) {
	prog := `

func f a:any...

	if (len a) == 0
	    print "no args"
	    return
	else if (typeof a[0]) == "num"
	    n := a[0].(num)
	    print "num" n
	else if (typeof a[0]) == "[]num"
	    na := a[0].([]num)
	    print "[]num" na
	else if (typeof a[0]) == "[]any"
	    aa := a[0].([]any)
	    print "[]any" aa
	else if (typeof a[0]) == "[][]any"
	    aa := a[0].([][]any)
	    print "[][]any" aa
	else
	    print "typeof:" (typeof a[0])
	end

end
f 1
f [1 2]
f []
f [[] {}]
f [[]]
f [{}]
f [[1]]
a:any
f a
`

	want := `
num 1
[]num [1 2]
[]any []
[]any [[] {}]
[][]any [[]]
typeof: []{}any
typeof: [][]num
typeof: bool
`[1:]

	got := run(prog)
	assert.Equal(t, want, got)
}

func TestAnyNestedArrayParam(t *testing.T) {
	prog := `
func f a:[][]any
    print "a:" a "typeof a[0]:" (typeof a[0])
    // a2 := a[0].([]any) // error: value of type assertion must be of type any, not []any
end
f [[]]
`
	got := run(prog)
	got = strings.TrimSpace(got)
	want := "a: [[]] typeof a[0]: []any"
	assert.Equal(t, want, got)
}

func TestAnyArrayParam(t *testing.T) {
	prog := `
func f a:[]any
    if (len a) == 0
        print "typeof a:" (typeof a) "empty a:" a
    else
        print "typeof a:" (typeof a) "typeof a[0]:" (typeof a[0]) "a:" a
    end
end
`
	tests := map[string]string{
		"f [1 2]":   "typeof a: []any typeof a[0]: num a: [1 2]",
		"f []":      "typeof a: []any empty a: []",
		"f [[] {}]": "typeof a: []any typeof a[0]: []any a: [[] {}]",
		"f [[]]":    "typeof a: []any typeof a[0]: []any a: [[]]",
		"f [{}]":    "typeof a: []any typeof a[0]: {}any a: [{}]",
		"f [[1]]":   "typeof a: []any typeof a[0]: []num a: [[1]]",
		`a:any
		f [a]`: "typeof a: []any typeof a[0]: bool a: [false]",
	}
	for input, want := range tests {
		got := run(prog + input)
		got = strings.TrimSpace(got)
		assert.Equal(t, want, got, input)
	}
}

func TestAnyArrayArrayParam(t *testing.T) {
	prog := `
func f a:[][]any
    if (len a) == 0
        print "1 typeof a:" (typeof a) "a:" a
    else if (len a[0]) == 0
        print "2 typeof a:" (typeof a) "typeof a[0]:" (typeof a[0]) "a:" a
    else
        print "3 typeof a:" (typeof a)  "typeof a[0]:" (typeof a[0]) "typeof a[0][0]:" (typeof a[0][0]) "a:" a
    end
end
`
	tests := map[string]string{
		"f [[1 2]]": "3 typeof a: [][]any typeof a[0]: []any typeof a[0][0]: num a: [[1 2]]",
		"f []":      "1 typeof a: [][]any a: []",
		"f [[] []]": "2 typeof a: [][]any typeof a[0]: []any a: [[] []]",
		`a:[][]any
		 f a`: "1 typeof a: [][]any a: []",
		`a:[]any
		 f [a]`: "2 typeof a: [][]any typeof a[0]: []any a: [[]]",
		`a:any
		 f [[a]]`: "3 typeof a: [][]any typeof a[0]: []any typeof a[0][0]: bool a: [[false]]",
	}
	for input, want := range tests {
		got := run(prog + input)
		got = strings.TrimSpace(got)
		assert.Equal(t, want, got, input)
	}
}

func TestAnyMapParam(t *testing.T) {
	prog := `
func f m:{}any
     if (len m) == 0
         print "typeof m:" (typeof m) "empty m:" m
    else
         print "typeof m:" (typeof m) "typeof m.a:" (typeof m.a) "m:" m
    end
end
`
	tests := map[string]string{
		"f {a:1 b:2}":   "typeof m: {}any typeof m.a: num m: {a:1 b:2}",
		"f {}":          "typeof m: {}any empty m: {}",
		"f {a:{} b:[]}": "typeof m: {}any typeof m.a: {}any m: {a:{} b:[]}",
		"f {a:{}}":      "typeof m: {}any typeof m.a: {}any m: {a:{}}",
		"f {a:[]}":      "typeof m: {}any typeof m.a: []any m: {a:[]}",
		"f {a:{a:1}}":   "typeof m: {}any typeof m.a: {}num m: {a:{a:1}}",
		`a:any
		f {a:a}`: "typeof m: {}any typeof m.a: bool m: {a:false}",
	}
	for input, want := range tests {
		got := run(prog + input)
		got = strings.TrimSpace(got)
		assert.Equal(t, want, got, input)
	}
}

func TestAnyMapMapParam(t *testing.T) {
	prog := `
func f m:{}{}any
     if (len m) == 0
         print "1 typeof m:" (typeof m) "m:" m
    else if (len m.a) == 0
         print "2 typeof m:" (typeof m) "typeof m.a:" (typeof m.a) "m:" m
    else
         print "3 typeof m:" (typeof m)  "typeof m.a:" (typeof m.a) "typeof m.a.a:" (typeof m.a.a) "m:" m
    end
end
`
	tests := map[string]string{
		"f {a:{a:1 b:2}}": "3 typeof m: {}{}any typeof m.a: {}any typeof m.a.a: num m: {a:{a:1 b:2}}",
		"f {}":            "1 typeof m: {}{}any m: {}",
		"f {a:{} b:{}}":   "2 typeof m: {}{}any typeof m.a: {}any m: {a:{} b:{}}",
		`a:{}{}any
		 f a`: "1 typeof m: {}{}any m: {}",
		`a:{}any
		 f {a:a}`: "2 typeof m: {}{}any typeof m.a: {}any m: {a:{}}",
		`a:any
		 f {a:{a:a}}`: "3 typeof m: {}{}any typeof m.a: {}any typeof m.a.a: bool m: {a:{a:false}}",
	}
	for input, want := range tests {
		got := run(prog + input)
		got = strings.TrimSpace(got)
		assert.Equal(t, want, got, input)
	}
}

func TestAnyReturnType(t *testing.T) {
	prog := `

func f1:any
	return 1
end
func f2:any
	a:num
	return a
end
func f3:any
	a:any
	return a
end
func f4:any
	return [1]
end
func f5:any
	return {}
end
print (f1) (f2) (f3) (f4) (f5)
`
	got := run(prog)
	got = strings.TrimSpace(got)
	assert.Equal(t, "1 0 false [1] {}", got)
}

func TestAnyCompositeReturnType(t *testing.T) {
	prog := `
func f1:[]any
    return [1]
end
func f2:[]any
    return [[]]
end
func f3:[][]any
    return [[[true]]]
end
func f4:{}any
    return {}
end
func f5:{}any
    return {A:{}}
end
func f6:{}{}any
    return {A:{A:true}}
end
func f7:{}{}any
    return {A:{A:{A:true}}}
end
print (f1) (f2) (f3) (f4) (f5) (f6) (f7)
`
	got := run(prog)
	got = strings.TrimSpace(got)
	assert.Equal(t, "[1] [[]] [[[true]]] {} {A:{}} {A:{A:true}} {A:{A:{A:true}}}", got)
}

func TestAnyCompositeVar(t *testing.T) {
	prog := `m:{}any
m = {a:true}
print (typeof m) m
m = {a:1 b:2 c:true}
`
	want := "{}any {a:true}\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestLateCompositeLiteralTyping(t *testing.T) {
	prog := `
a:[]any
m:{}any
a = [1 2 3]
m = {a:true b:false}
print "a:" a (typeof a) "m:" m (typeof m)
`
	want := "a: [1 2 3] []any m: {a:true b:false} {}any\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestTypeofNestedArray(t *testing.T) {
	prog := `a := [[]]
print (typeof a)
print (typeof a[0])`

	want := "[][]any\n[]any\n"
	got := run(prog)
	assert.Equal(t, want, got)
}

func TestNestedTypeof(t *testing.T) {
	tests := map[string]string{
		"a := [[] [1] []]":       "[][]num",
		"a := [[[]] [[1]] [[]]]": "[][][]num",
		"a := [[]]":              "[][]any",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			in += "\n print (typeof a)"
			var got string
			rt := &testRT{}
			rt.UnimplementedRuntime.print = rt.Print
			eval := NewEvaluator(rt)
			err := eval.Run(in)
			if err != nil {
				got = err.Error()
			} else {
				got = rt.b.String()
			}
			assert.Equal(t, want+"\n", got)
		})
	}
}

func TestAnyEquals(t *testing.T) {
	prog := `
a:any
b:any

a = 1
b = "hello"
print 1 (a == b) (a != b)

a = 1
b = 2
print 2 (a == b) (a != b)

a = 1
b = 1
print 3 (a == b) (a != b)
`
	want := `
1 false true
2 false true
3 true false
`[1:]
	got := run(prog)
	assert.Equal(t, want, got)
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
	want := `
"move" not implemented
"line" not implemented
x: 12
🍦 big x
`[1:]
	got := run(prog)
	assert.Equal(t, want, got)
}
