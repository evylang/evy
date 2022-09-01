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
		"a:num[]":     []string{"a:ARRAY NUM=[]"},
		"a:num[]{}":   []string{"a:MAP ARRAY NUM={}"},
		"abc:any[]{}": []string{"abc:MAP ARRAY ANY={}"},
		//TODO: array lit etc. "a := num[]":  []string{"a:ARRAY NUM"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input)
		got := parser.Parse()
		assert.Equal(t, 0, len(parser.errors), "input: %s\nerrors: %#v", input, parser.errors)
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
		assert.Equal(t, 0, len(parser.errors), "input: %s\nerrors: %#v", input, parser.errors)
		assert.Equal(t, "\n", got.String())
	}
}

func TestParseDeclarationError(t *testing.T) {
	tests := map[string][]string{
		"a :invalid":    []string{"a:ILLEGAL_TYPE"},
		"a :":           []string{"a:ILLEGAL_TYPE"},
		"a :\n":         []string{"a:ILLEGAL_TYPE"},
		"a ://blabla\n": []string{"a:ILLEGAL_TYPE"},
		"a :true":       []string{"a:ILLEGAL_TYPE"},
		"a :[]":         []string{"a:ILLEGAL_TYPE"},
		"a :num num":    []string{"a:ILLEGAL_TYPE"},
		"a :[]num":      []string{"a:ILLEGAL_TYPE"},
		"a :()":         []string{"a:ILLEGAL_TYPE"},
		"a :num{}num":   []string{"a:ILLEGAL_TYPE"},
		"a ::":          []string{"a:ILLEGAL_TYPE"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input)
		got := parser.Parse()
		assert.Equal(t, 1, len(parser.errors), "input: %s\nerrors: %#v", input, parser.errors)
		assert.Equal(t, want, got.String())
	}
}

func TestFunctionCall(t *testing.T) {
	tests := map[string][]string{
		"print":               []string{"print()"},
		"print 123":           []string{"print(123)"},
		`print 123 "abc"`:     []string{"print(123, 'abc')"},
		"a:=1 \n print a":     []string{"a:NUM=1", "print(a:NUM)"},
		"a:string \n print a": []string{"a:STRING=''", "print(a:STRING)"},
		`a:=true
		b:string
		print a b`: []string{"a:BOOL=true", "b:STRING=''", "print(a:BOOL, b:STRING)"},
	}
	for input, wantSlice := range tests {
		want := strings.Join(wantSlice, "\n") + "\n"
		parser := New(input)
		got := parser.Parse()
		assert.Equal(t, 0, len(parser.errors), "input: %s\nerrors: %#v", input, parser.errors)
		assert.Equal(t, want, got.String())
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
		assert.Equal(t, 0, len(parser.errors), "input: %s\nerrors: %#v", input, parser.errors)
		assert.Equal(t, want, got.String())
	}
}

func TestDemo(t *testing.T) {
	input := `
move 10 10
line 20 20
x := 12
if x > 10
	print "🍦 big x" x
end`
	parser := New(input)
	got := parser.Parse()
	assert.Equal(t, 2, len(parser.errors))
	assert.Equal(t, "line 2 column 1: unknown function 'move'", parser.errors[0].String())
	assert.Equal(t, "line 3 column 1: unknown function 'line'", parser.errors[1].String())
	want := "x:NUM=12\n"
	assert.Equal(t, want, got.String())
}
