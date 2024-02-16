package bytecode

import (
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
			name: "addition and subtraction",
			input: `
			x := 2 - 1 + 3
			x = x
			`,
			expectedStackTop:  4,
			expectedConstants: []any{2, 1, 3},
			expectedInstructions: []Instructions{
				mustMake(t, OpConstant, 0),
				mustMake(t, OpConstant, 1),
				mustMake(t, OpSubtract),
				mustMake(t, OpConstant, 2),
				mustMake(t, OpAdd),
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
	assert.Equal(t, len(concatted), len(actual), "wrong instructions length")
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
