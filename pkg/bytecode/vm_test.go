package bytecode

import (
	"math"
	"testing"

	"evylang.dev/evy/pkg/assert"
	"evylang.dev/evy/pkg/parser"
)

func TestVMGlobals(t *testing.T) {
	tests := []testCase{
		{
			name: "global assignment",
			input: `
			x := 1
			x = x
			`,
			expectedStackTop:  1,
			expectedConstants: []any{1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input, parser.Builtins{})
			assert.NoError(t, err, "parser error")
			comp := NewCompiler()
			err = comp.Compile(program)
			assert.NoError(t, err, "compiler error")
			vm := NewVM(comp.Bytecode())
			err = vm.Run()
			assert.NoError(t, err, "runtime error")
			stackElem := vm.lastPoppedStackElem()
			switch expected := tt.expectedStackTop.(type) {
			case int:
				assertNumValue(t, float64(expected), stackElem)
			case float64:
				assertNumValue(t, expected, stackElem)
			default:
				t.Errorf("unexpected object type %v", expected)
			}
		})
	}
}

func TestUserError(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name: "divide by zero",
			input: `
			x := 2 / 0
			x = x
			`,
			expectedErr: ErrDivideByZero,
		},
		{
			name: "modulo by zero",
			input: `
			x := 2 % 0
			x = x
			`,
			expectedErr: ErrDivideByZero,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input, parser.Builtins{})
			assert.NoError(t, err, "parser error")
			comp := NewCompiler()
			err = comp.Compile(program)
			assert.NoError(t, err, "compiler error")
			vm := NewVM(comp.Bytecode())
			err = vm.Run()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestVMArithmetic(t *testing.T) {
	tests := []testCase{
		{
			name: "addition",
			input: `
			x := 2 + 1
			x = x
			`,
			expectedStackTop:  3,
			expectedConstants: []any{2, 1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpAdd),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "subtraction",
			input: `
			x := 2 - 1
			x = x
			`,
			expectedStackTop:  1,
			expectedConstants: []any{2, 1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpSubtract),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "multiplication",
			input: `
			x := 2 * 1
			x = x
			`,
			expectedStackTop:  2,
			expectedConstants: []any{2, 1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpMultiply),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "division",
			input: `
			x := 2 / 1
			x = x
			`,
			expectedStackTop:  2,
			expectedConstants: []any{2, 1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpDivide),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "modulo",
			input: `
			x := 2 % 1
			x = x
			`,
			expectedStackTop:  0,
			expectedConstants: []any{2, 1},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpModulo),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "float modulo",
			input: `
			x := 2.5 % 1.3
			x = x
			`,
			expectedStackTop:  1.2,
			expectedConstants: []any{2.5, 1.3},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpModulo),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "all operators",
			input: `
			x := 1 + 2 - 3 * 4 / 5 % 6
			x = x
			`,
			expectedStackTop:  1 + 2 - math.Mod(3.0*4.0/5.0, 6.0),
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpAdd),
				mustMake(t, OpConstant, 2),
				mustMake(t, OpConstant, 3),
				mustMake(t, OpMultiply),
				mustMake(t, OpConstant, 4),
				mustMake(t, OpDivide),
				mustMake(t, OpConstant, 5),
				mustMake(t, OpModulo),
				mustMake(t, OpSubtract),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
		{
			name: "grouped expressions",
			input: `
			x := (1 + 2 - 3) * 4 / 5 % 6
			x = x
			`,
			expectedStackTop:  (1 + 2 - 3) * 4 / math.Mod(5.0, 6.0),
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpAdd),
				mustMake(t, OpConstant, 2),
				mustMake(t, OpSubtract),
				mustMake(t, OpConstant, 3),
				mustMake(t, OpMultiply),
				mustMake(t, OpConstant, 4),
				mustMake(t, OpDivide),
				mustMake(t, OpConstant, 5),
				mustMake(t, OpModulo),
				mustMake(t, OpSetGlobal, 0),
				mustMake(t, OpGetGlobal, 0),
				mustMake(t, OpSetGlobal, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input, parser.Builtins{})
			assert.NoError(t, err, "parser error")
			comp := NewCompiler()
			err = comp.Compile(program)
			assert.NoError(t, err, "compiler error")
			bytecode := comp.Bytecode()
			assertInstructions(t, tt.expectedInstructions, bytecode.Instructions)
			assertConstants(t, tt.expectedConstants, bytecode.Constants)
			vm := NewVM(bytecode)
			err = vm.Run()
			assert.NoError(t, err, "runtime error")
			stackElem := vm.lastPoppedStackElem()
			switch expected := tt.expectedStackTop.(type) {
			case int:
				assertNumValue(t, float64(expected), stackElem)
			case float64:
				assertNumValue(t, expected, stackElem)
			default:
				t.Errorf("unexpected object type %v", expected)
			}
		})
	}
}

// testCase covers both the compiler and the VM.
type testCase struct {
	name string
	// input is an evy program
	input string
	// expectedStackTop is the result of popping the last
	// element from the stack in the vm.
	expectedStackTop any
	// expectedConstants are the expected constants passed in the
	// bytecode after compilation.
	expectedConstants []any
	// expectedInstructions are the expected compiler instructions
	// passed in the bytecode after compilation
	expectedInstructions []Instructions
}

func assertInstructions(t *testing.T, expected []Instructions, actual Instructions) {
	t.Helper()
	concatted := concatInstructions(expected)
	assert.Equal(t, len(concatted), len(actual), "wrong instructions length\nwant=\n%s\ngot=\n%s",
		concatted, actual)
	for i, ins := range concatted {
		assert.Equal(t, ins, actual[i], "wrong instruction at %04d\nwant=\n%s\ngot=\n%s",
			i, concatted, actual)
	}
}

func concatInstructions(s []Instructions) Instructions {
	out := Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func assertConstants(t *testing.T, expected []any, actual []value) {
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
	result, ok := actual.(numVal)
	assert.Equal(t, true, ok, "object is not a NumVal. got=%T (%+v)", actual, actual)
	assert.Equal(t, expected, float64(result), "object has wrong value")
}
