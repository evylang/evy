package evaluator

import (
	"bytes"
	"testing"

	"foxygo.at/evy/pkg/assert"
)

func TestParseDeclaration(t *testing.T) {
	tests := map[string]string{
		"a:=1\n print a 2": "1 2",
	}
	for in, want := range tests {
		in, want := in, want
		t.Run(in, func(t *testing.T) {
			b := bytes.Buffer{}
			fn := func(s string) { b.WriteString(s) }
			Run(in, fn)
			assert.Equal(t, want, b.String())
		})
	}
}

func TestDemo(t *testing.T) {
	prog := `move 10 10
line 20 20
x := 12
if x > 10
    print "🍦 big x" x
end`
	b := bytes.Buffer{}
	fn := func(s string) { b.WriteString(s) }
	Run(prog, fn)
	want := ""
	assert.Equal(t, want, b.String())
}
