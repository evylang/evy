package bytecode

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
	"evylang.dev/evy/pkg/parser"
)

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"x := 1\nx = x", 1},
		{"y := 2\ny = y", 2},
	}
	for _, tt := range tests {
		program, err := parser.Parse(tt.input, parser.Builtins{})
		assert.NoError(t, err, "parser error")
		comp := NewCompiler()
		err = comp.Compile(program)
		assert.NoError(t, err, "compiler error")
		vm := NewVM(comp.Bytecode())
		err = vm.Run()
		assert.NoError(t, err, "vm error")
		stackElem := vm.lastPoppedStackElem()
		switch expected := tt.expected.(type) {
		case int:
			assertNumValue(t, float64(expected), stackElem)
		case float64:
			assertNumValue(t, expected, stackElem)
		default:
			t.Errorf("unexpected object type %v", expected)
		}
	}
}

type vmTestCase struct {
	input    string
	expected interface{}
}
