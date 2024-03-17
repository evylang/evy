package bytecode

import (
	"math"
	"testing"

	"evylang.dev/evy/pkg/assert"
	"evylang.dev/evy/pkg/parser"
)

// testCase covers both the compiler and the VM.
type testCase struct {
	name string
	// input is an evy program
	input string
	// wantStackTop is the expected last popped element of the stack in the vm.
	wantStackTop value
	// wantBytecode is the bytecode expected from the compiler.
	wantBytecode *Bytecode
}

func TestVMGlobals(t *testing.T) {
	tests := []testCase{
		{
			name:         "global assignment",
			input:        "x := 1",
			wantStackTop: makeValue(t, 1),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			assertBytecode(t, tt.wantBytecode, bytecode)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.NoError(t, err, "runtime error")

			got := vm.lastPoppedStackElem()
			assert.Equal(t, tt.wantStackTop, got)
		})
	}
}

func TestUserError(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "divide by zero",
			input:   "x := 2 / 0",
			wantErr: ErrDivideByZero,
		}, {
			name:    "modulo by zero",
			input:   "x := 2 % 0",
			wantErr: ErrDivideByZero,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)

			vm := NewVM(bytecode)
			err := vm.Run()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestBoolExpressions(t *testing.T) {
	tests := []testCase{
		{
			name:         "literal true",
			input:        "x := true",
			wantStackTop: makeValue(t, true),
			wantBytecode: &Bytecode{
				Instructions: makeInstructions(
					mustMake(t, OpTrue),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "literal false",
			input:        "x := false",
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Instructions: makeInstructions(
					mustMake(t, OpFalse),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "not operator",
			input:        "x := !true",
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Instructions: makeInstructions(
					mustMake(t, OpTrue),
					mustMake(t, OpNot),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "equal operator",
			input:        "x := 1 == 1",
			wantStackTop: makeValue(t, true),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpEqual),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "not operator",
			input:        "x := 1 != 1",
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNotEqual),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			assertBytecode(t, tt.wantBytecode, bytecode)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.NoError(t, err, "runtime error")

			got := vm.lastPoppedStackElem()
			assert.Equal(t, tt.wantStackTop, got)
		})
	}
}

func TestVMArithmetic(t *testing.T) {
	tests := []testCase{
		{
			name:         "addition",
			input:        "x := 2 + 1",
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 2, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpAdd),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "subtraction",
			input:        "x := 2 - 1",
			wantStackTop: makeValue(t, 1),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 2, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpSubtract),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "multiplication",
			input:        "x := 2 * 1",
			wantStackTop: makeValue(t, 2),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 2, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpMultiply),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "division",
			input:        "x := 2 / 1",
			wantStackTop: makeValue(t, 2),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 2, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpDivide),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "modulo",
			input:        "x := 2 % 1",
			wantStackTop: makeValue(t, 0),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 2, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpModulo),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "float modulo",
			input:        "x := 2.5 % 1.3",
			wantStackTop: makeValue(t, 1.2),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 2.5, 1.3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpModulo),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "minus operator",
			input:        "x := -1",
			wantStackTop: makeValue(t, -1),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpMinus),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		}, {
			name:         "all operators",
			input:        "x := 1 + 2 - 3 * 4 / 5 % 6",
			wantStackTop: makeValue(t, 1+2-math.Mod(3.0*4.0/5.0, 6.0)),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 4, 5, 6),
				Instructions: makeInstructions(
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
				),
			},
		}, {
			name:         "grouped expressions",
			input:        "x := (1 + 2 - 3) * 4 / 5 % 6",
			wantStackTop: makeValue(t, (1+2-3)*4/math.Mod(5.0, 6.0)),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 4, 5, 6),
				Instructions: makeInstructions(
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
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			assertBytecode(t, tt.wantBytecode, bytecode)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.NoError(t, err, "runtime error")
			got := vm.lastPoppedStackElem()
			assert.Equal(t, tt.wantStackTop, got)
		})
	}
}

func compileBytecode(t *testing.T, input string) *Bytecode {
	t.Helper()
	// add x = x to the input so it parses correctly
	input += "\nx = x"
	program, err := parser.Parse(input, parser.Builtins{})
	assert.NoError(t, err, "unexpected parse error")
	comp := NewCompiler()
	err = comp.Compile(program)
	assert.NoError(t, err, "unexpected compile error")
	bc := comp.Bytecode()
	// remove the final 2 instructions that represent x = x
	bc.Instructions = bc.Instructions[:len(bc.Instructions)-6]
	return bc
}

func makeValue(t *testing.T, a any) value {
	t.Helper()
	switch v := a.(type) {
	case int:
		return numVal(v)
	case float64:
		return numVal(v)
	case bool:
		return boolVal(v)
	default:
		t.Fatalf("makeValue(%q): invalid type: %T", a, a)
		return nil
	}
}

func makeValues(t *testing.T, in ...any) []value {
	t.Helper()
	out := make([]value, len(in))
	for i, a := range in {
		out[i] = makeValue(t, a)
	}
	return out
}

func makeInstructions(in ...Instructions) Instructions {
	out := Instructions{}
	for _, instruction := range in {
		out = append(out, instruction...)
	}
	return out
}

func assertBytecode(t *testing.T, want, got *Bytecode) {
	t.Helper()
	msg := "\nwant instructions:\n%s\ngot instructions:\n%s"
	assert.Equal(t, want, got, msg, want.Instructions, got.Instructions)
}
