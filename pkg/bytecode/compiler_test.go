package bytecode

import (
	"testing"

	"evylang.dev/evy/pkg/assert"
	"evylang.dev/evy/pkg/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []Instructions
}

func TestGlobalVarStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "x := 1\nx = x",
			expectedConstants: []interface{}{1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			input:             "x := 2\nx = x",
			expectedConstants: []interface{}{2},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
	}
	for _, tt := range tests {
		program, err := parser.Parse(tt.input, parser.Builtins{})
		assert.NoError(t, err, "parser error")
		compiler := NewCompiler()
		err = compiler.Compile(program)
		assert.NoError(t, err, "compiler error")
		bytecode := compiler.Bytecode()
		assertInstructions(t, tt.expectedInstructions, bytecode.Instructions)
		assertConstants(t, tt.expectedConstants, bytecode.Constants)
	}
}

func assertInstructions(t *testing.T, expected []Instructions, actual Instructions) {
	t.Helper()
	concatted := concatInstructions(expected)
	assert.Equal(t, len(concatted), len(actual), "wrong instructions length")
	for i, ins := range concatted {
		assert.Equal(t, ins, actual[i], "wrong instruction %d", i)
	}
}

func concatInstructions(s []Instructions) Instructions {
	out := Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func assertConstants(t *testing.T, expected []interface{}, actual []value) {
	t.Helper()
	assert.Equal(t, len(expected), len(actual), "wrong number of constants")
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			assertNumValue(t, float64(constant), actual[i])
		case float64:
			assertNumValue(t, constant, actual[i])
		default:
			t.Errorf("unknown constant type %v", constant)
		}
	}
}

func assertNumValue(t *testing.T, expected float64, actual value) {
	t.Helper()
	result, ok := actual.(*numVal)
	assert.Equal(t, true, ok, "object is not a NumVal. got=%T (%+v)", actual, actual)
	assert.Equal(t, expected, result.V, "object has wrong value")
}
