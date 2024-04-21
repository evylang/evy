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
			name:         "inferred declaration",
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
		{
			name: "assignment",
			input: `x := 1
			y := x
			y = x + 1
			y = y`,
			wantStackTop: makeValue(t, 2),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 1),
				Instructions: makeInstructions(
					// x := 1
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					// y := x
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpSetGlobal, 1),
					// y = x + 1
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpAdd),
					mustMake(t, OpSetGlobal, 1),
					// y = y
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpSetGlobal, 1),
				),
			},
		},
		{
			name: "index assignment",
			input: `x := [1 2 3]
			x[0] = x[2]
			x[2] = 1
			x = x`,
			wantStackTop: makeValue(t, []any{3, 2, 1}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 2, 0, 1, 2),
				Instructions: makeInstructions(
					// x := [1 2 3]
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpSetGlobal, 0),
					// x[2]
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpIndex),
					// x[0]
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpSetIndex),
					// 1
					mustMake(t, OpConstant, 5),
					// x[2]
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 6),
					mustMake(t, OpSetIndex),
					// x = x
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "nested index assignment",
			input: `x := [[1 2] [3 4]]
			x[0][0] = x[0][1]
			x = x`,
			wantStackTop: makeValue(t, []any{[]any{2, 2}, []any{3, 4}}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 4, 0, 1, 0, 0),
				Instructions: makeInstructions(
					// x := [[1 2] [3 4]]
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpArray, 2),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpArray, 2),
					mustMake(t, OpArray, 2),
					mustMake(t, OpSetGlobal, 0),
					// x[0][0] = x[0][1]
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpIndex),
					mustMake(t, OpConstant, 5),
					mustMake(t, OpIndex),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 6),
					mustMake(t, OpIndex),
					mustMake(t, OpConstant, 7),
					mustMake(t, OpSetIndex),
					// x = x
					mustMake(t, OpGetGlobal, 0),
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

func TestNumOperations(t *testing.T) {
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		{
			name:         "less than",
			input:        "x := 1 < 2",
			wantStackTop: makeValue(t, true),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumLessThan),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "less than or equal",
			input:        "x := 1 <= 2",
			wantStackTop: makeValue(t, true),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumLessThanEqual),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "greater than",
			input:        "x := 1 > 2",
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumGreaterThan),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "greater than or equal",
			input:        "x := 1 >= 2",
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumGreaterThanEqual),
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

func TestStringExpressions(t *testing.T) {
	tests := []testCase{
		{
			name:         "assignment",
			input:        `x := "a"`,
			wantStackTop: makeValue(t, "a"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "concatenate",
			input:        `x := "a" + "b"`,
			wantStackTop: makeValue(t, "ab"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", "b"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpStringConcatenate),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "less than",
			input:        `x := "a" < "b"`,
			wantStackTop: makeValue(t, true),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", "b"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpStringLessThan),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "less than or equal",
			input:        `x := "a" <= "b"`,
			wantStackTop: makeValue(t, true),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", "b"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpStringLessThanEqual),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "greater than",
			input:        `x := "a" > "b"`,
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", "b"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpStringGreaterThan),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "greater than or equal",
			input:        `x := "a" >= "b"`,
			wantStackTop: makeValue(t, false),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", "b"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpStringGreaterThanEqual),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "index",
			input:        `x := "abc"[0]`,
			wantStackTop: makeValue(t, "a"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 0),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "negative index",
			input:        `x := "abc"[-1]`,
			wantStackTop: makeValue(t, "c"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpMinus),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice",
			input:        `x := "abc"[1:3]`,
			wantStackTop: makeValue(t, "bc"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 1, 3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "negative slice",
			input:        `x := "abc"[-3:-1]`,
			wantStackTop: makeValue(t, "ab"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 3, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpMinus),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpMinus),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice default start",
			input:        `x := "abc"[:1]`,
			wantStackTop: makeValue(t, "a"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpNone),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice default end",
			input:        `x := "abc"[1:]`,
			wantStackTop: makeValue(t, "bc"),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNone),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice start equals length",
			input:        `x := "abc"[3:]`,
			wantStackTop: makeValue(t, ""),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "abc", 3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNone),
					mustMake(t, OpSlice),
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

//nolint:maintidx // This function is not complex by any measure.
func TestArrays(t *testing.T) {
	tests := []testCase{
		{
			name:         "empty assignment",
			input:        "x := []",
			wantStackTop: makeValue(t, []any{}),
			wantBytecode: &Bytecode{
				Constants: nil,
				Instructions: makeInstructions(
					mustMake(t, OpArray, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "assignment",
			input:        "x := [1 2 3]",
			wantStackTop: makeValue(t, []any{1, 2, 3}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "concatenate",
			input:        "x := [1 2] + [3 4]",
			wantStackTop: makeValue(t, []any{1, 2, 3, 4}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 4),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpArray, 2),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpArray, 2),
					mustMake(t, OpArrayConcatenate),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "repetition",
			input:        `x := [1 2] * 3`,
			wantStackTop: makeValue(t, []any{1, 2, 1, 2, 1, 2}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpArray, 2),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArrayRepeat),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "zero repetition",
			input:        `x := [1 2] * 0`,
			wantStackTop: makeValue(t, []any{}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 0),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpArray, 2),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArrayRepeat),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "concatenate preserve original",
			input: `x := [1 2]
			y := [3 4]
			y = x + y
			x = x`,
			wantStackTop: makeValue(t, []any{1, 2}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 4),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpArray, 2),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpArray, 2),
					mustMake(t, OpSetGlobal, 1),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpArrayConcatenate),
					mustMake(t, OpSetGlobal, 1),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "index",
			input:        `x := [1 2 3][0]`,
			wantStackTop: makeValue(t, 1),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 0),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "negative index",
			input:        `x := [1 2 3][-1]`,
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpMinus),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice",
			input:        `x := [1 2 3][1:3]`,
			wantStackTop: makeValue(t, []any{2, 3}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 1, 3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "negative slice",
			input:        `x := [1 2 3][-3:-1]`,
			wantStackTop: makeValue(t, []any{1, 2}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 3, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpMinus),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpMinus),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice default start",
			input:        `x := [1 2 3][:1]`,
			wantStackTop: makeValue(t, []any{1}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpNone),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice default end",
			input:        `x := [1 2 3][1:]`,
			wantStackTop: makeValue(t, []any{2, 3}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpNone),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "slice start equals length",
			input:        `x := [1 2 3][3:]`,
			wantStackTop: makeValue(t, []any{}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 3),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpNone),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "slice preserve original",
			input: `x := [1 2 3]
			y := x[1:]
			y[0] = 8
			x = x`,
			wantStackTop: makeValue(t, []any{1, 2, 3}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 1, 8, 0),
				Instructions: makeInstructions(
					// x := [1 2 3]
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpArray, 3),
					mustMake(t, OpSetGlobal, 0),
					// y := x[1:]
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpNone),
					mustMake(t, OpSlice),
					mustMake(t, OpSetGlobal, 1),
					// y[0] = 8
					mustMake(t, OpConstant, 4),
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpConstant, 5),
					mustMake(t, OpSetIndex),
					// x = x
					mustMake(t, OpGetGlobal, 0),
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

func TestErrArrayRepetition(t *testing.T) {
	type repeatTest struct {
		name  string
		input string
	}
	tests := []repeatTest{
		{
			name:  "non-integer repetition",
			input: `x := [1 2 3] * 4.5`,
		},
		{
			name:  "negative repetition",
			input: `x := [1 2 3] * -1`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.Error(t, ErrBadRepetition, err)
		})
	}
}

func TestErrBounds(t *testing.T) {
	type boundsTest struct {
		name  string
		input string
	}
	tests := []boundsTest{
		{
			name:  "string index out of bounds",
			input: `x := "abc"[3]`,
		},
		{
			name:  "string negative index out of bounds",
			input: `x := "abc"[-4]`,
		},
		{
			name:  "array index out of bounds",
			input: `x := [1 2 3][3]`,
		},
		{
			name:  "array negative index out of bounds",
			input: `x := [1 2 3][-4]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.Error(t, ErrBounds, err)
		})
	}
}

func TestErrIndex(t *testing.T) {
	type boundsTest struct {
		name  string
		input string
	}
	tests := []boundsTest{
		{
			name:  "string index not an integer",
			input: `x := "abc"[1.1]`,
		},
		{
			name:  "array index not an integer",
			input: `x := [1 2 3][1.1]`,
		},
		{
			name:  "array invalid slice type",
			input: `x := [1 2 3][1.1:2.1]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.Error(t, ErrIndexValue, err)
		})
	}
}

func TestErrSlice(t *testing.T) {
	type boundsTest struct {
		name  string
		input string
	}
	tests := []boundsTest{
		{
			name:  "string invalid slice",
			input: `x := "abc"[2:1]`,
		},
		{
			name:  "array invalid slice",
			input: `x := [1 2 3][2:1]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.Error(t, ErrSlice, err)
		})
	}
}

func TestMap(t *testing.T) {
	tests := []testCase{
		{
			name:         "empty",
			input:        "x := {}",
			wantStackTop: makeValue(t, map[string]any{}),
			wantBytecode: &Bytecode{
				Constants: nil,
				Instructions: makeInstructions(
					mustMake(t, OpMap, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "assignment",
			input:        "x := {a: 1 b: 2}",
			wantStackTop: makeValue(t, map[string]any{"a": 1, "b": 2}),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", 1, "b", 2),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpMap, 4),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name:         "index",
			input:        `x := {a: 1 b: 2}["b"]`,
			wantStackTop: makeValue(t, 2),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, "a", 1, "b", 2, "b"),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpMap, 4),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpIndex),
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

func TestErrMapKey(t *testing.T) {
	input := `x := {a: 1}["b"]`
	bytecode := compileBytecode(t, input)
	vm := NewVM(bytecode)
	err := vm.Run()
	assert.Error(t, ErrMapKey, err)
}

func TestIf(t *testing.T) {
	tests := []testCase{
		{
			name: "single if",
			input: `x := 1
			if x == 1
				x = 2
			end`,
			wantStackTop: makeValue(t, 2),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 1, 2),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpEqual),
					mustMake(t, OpJumpOnFalse, 25),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 25),
				),
			},
		},
		{
			name: "if else",
			input: `x := 10
			if x < 5
				x = 20
			else
				x = 5
			end`,
			wantStackTop: makeValue(t, 5),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 10, 5, 20, 5),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumLessThan),
					mustMake(t, OpJumpOnFalse, 25),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 31),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "else if else",
			input: `x := 10
			if x > 10
				x = 10
			else if x > 5
				x = 5
			else
				x = 1
			end`,
			wantStackTop: makeValue(t, 5),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 10, 10, 10, 5, 5, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumGreaterThan),
					mustMake(t, OpJumpOnFalse, 25),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 50),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpNumGreaterThan),
					mustMake(t, OpJumpOnFalse, 44),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 50),
					mustMake(t, OpConstant, 5),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "many elseif",
			input: `x := 3
				if x == 1
					x = 11
				else if x == 2
					x = 12
				else if x == 3
					x = 13
				else 
					x = 14
				end`,
			wantStackTop: makeValue(t, 13),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 3, 1, 11, 2, 12, 3, 13, 14),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpEqual),
					mustMake(t, OpJumpOnFalse, 25),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 69),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpEqual),
					mustMake(t, OpJumpOnFalse, 44),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 69),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 5),
					mustMake(t, OpEqual),
					mustMake(t, OpJumpOnFalse, 63),
					mustMake(t, OpConstant, 6),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 69),
					mustMake(t, OpConstant, 7),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "no matching if",
			input: `x := 1
			if false
				x = 2
			else if false
				x = 3
			else if false
				x = 4
			end
			x = x`,
			wantStackTop: makeValue(t, 1),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 1, 2, 3, 4),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpFalse),
					mustMake(t, OpJumpOnFalse, 19),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 45),
					mustMake(t, OpFalse),
					mustMake(t, OpJumpOnFalse, 32),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 45),
					mustMake(t, OpFalse),
					mustMake(t, OpJumpOnFalse, 45),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 45),
					mustMake(t, OpGetGlobal, 0),
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

func TestWhile(t *testing.T) {
	tests := []testCase{
		{
			name: "enter loop",
			input: `x := 0
			while x < 5
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 5),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 5, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumLessThan),
					mustMake(t, OpJumpOnFalse, 29),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpAdd),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 6),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "skip loop",
			input: `x := 0
			while x > 5
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 0),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 5, 1),
				Instructions: makeInstructions(
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumGreaterThan),
					mustMake(t, OpJumpOnFalse, 29),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpAdd),
					mustMake(t, OpSetGlobal, 0),
					mustMake(t, OpJump, 6),
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "break loop",
			input: `x := 0
			while x < 5
				x = x + 1
				if x == 3 
					break
				end
			end
			x = x`,
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 5, 1, 3),
				Instructions: makeInstructions(
					// x := 0
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					// while x < 5
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpNumLessThan),
					mustMake(t, OpJumpOnFalse, 45),
					// 		x = x + 1
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpAdd),
					mustMake(t, OpSetGlobal, 0),
					// 		if x == 3
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpEqual),
					mustMake(t, OpJumpOnFalse, 42),
					// 			break
					mustMake(t, OpJump, 45),
					// 		end
					mustMake(t, OpJump, 42),
					// end
					mustMake(t, OpJump, 6),
					// x = x
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpSetGlobal, 0),
				),
			},
		},
		{
			name: "nested loops and breaks",
			input: `
			x := 0
			while true
				while true
					break
				end
				x = x + 1
				break
			end
			x = x`,
			wantStackTop: makeValue(t, 1),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 1),
				Instructions: makeInstructions(
					// x := 0
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					// while true
					mustMake(t, OpTrue),
					mustMake(t, OpJumpOnFalse, 36),
					// 		while true
					mustMake(t, OpTrue),
					mustMake(t, OpJumpOnFalse, 20),
					// 			break
					mustMake(t, OpJump, 20),
					// 		end
					mustMake(t, OpJump, 10),
					// 		x = x + 1
					mustMake(t, OpGetGlobal, 0),
					mustMake(t, OpConstant, 1),
					mustMake(t, OpAdd),
					mustMake(t, OpSetGlobal, 0),
					// 		break
					mustMake(t, OpJump, 36),
					// end
					mustMake(t, OpJump, 6),
					// x = x
					mustMake(t, OpGetGlobal, 0),
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

func TestStepRange(t *testing.T) {
	tests := []testCase{
		{
			name: "for range default start and step",
			input: `x := 0
			for range 10
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 10),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
		{
			name: "for range default step",
			input: `x := 0
			for range 2 10
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 8),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
		{
			name: "for range var",
			input: `x := 0
			for i := range 10
				x = i
			end
			x = x`,
			wantStackTop: makeValue(t, 9),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
		{
			name: "for range step",
			input: `x := 0
			for i := range 0 10 4
				x = i
			end
			x = x`,
			wantStackTop: makeValue(t, 8),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
		{
			name: "for range negative step",
			input: `x := 0
			for i := range 10 0 -1
				x = i
			end
			x = x`,
			wantStackTop: makeValue(t, 1),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
		{
			name: "for range invalid stop",
			input: `x := 0
			for range -10
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 0),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
		{
			name: "for break",
			input: `x := 0
			for x := range 5
				if x == 3 
                    break
                end
			end
			x = x`,
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants:    makeValues(t),
				Instructions: makeInstructions(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			t.Log(bytecode.Constants)
			t.Log(bytecode.Instructions)
			// assertBytecode(t, tt.wantBytecode, bytecode)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.NoError(t, err, "runtime error")
			got := vm.lastPoppedStackElem()
			assert.Equal(t, tt.wantStackTop, got)
		})
	}
}

func TestArrayRange(t *testing.T) {
	tests := []testCase{
		{
			name: "for range array",
			input: `x := 0
			for e := range [1 2 3]
				x = e
			end
			x = x`,
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 0, 1, 2, 3),
				Instructions: makeInstructions(
					// x := 0
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					// i := 0 // i is the hidden index variable
					mustMake(t, OpConstant, 1),
					mustMake(t, OpSetGlobal, 1),
					// [1 2 3]
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpArray, 3),
					// e := arr[i]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 2),
					// [1 2 3]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpArray, 3),
					// e := arr[i]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpGetGlobal, 2),
				),
			},
		},
		{
			name: "for range array no loopvar",
			input: `x := 0
			for range [1 2 3]
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 0, 1, 2, 3),
				Instructions: makeInstructions(
					// x := 0
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					// i := 0
					mustMake(t, OpConstant, 1),
					mustMake(t, OpSetGlobal, 1),
					// [1 2 3]
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpArray, 3),
					// e := arr[i]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 2),
					// [1 2 3]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpArray, 3),
					// e := arr[i]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpGetGlobal, 2),
				),
			},
		},
		{
			name: "for range array variable",
			input: `x := 0
			y := [1 2 3]
			for e := range y
				x = e
			end
			x = x`,
			wantStackTop: makeValue(t, 3),
			wantBytecode: &Bytecode{
				Constants: makeValues(t, 0, 0, 1, 2, 3),
				Instructions: makeInstructions(
					// x := 0
					mustMake(t, OpConstant, 0),
					mustMake(t, OpSetGlobal, 0),
					// i := 0
					mustMake(t, OpConstant, 1),
					mustMake(t, OpSetGlobal, 1),
					// [1 2 3]
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpArray, 3),
					// e := arr[i]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpSetGlobal, 2),
					// [1 2 3]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpConstant, 2),
					mustMake(t, OpConstant, 3),
					mustMake(t, OpConstant, 4),
					mustMake(t, OpArray, 3),
					// e := arr[i]
					mustMake(t, OpGetGlobal, 1),
					mustMake(t, OpIndex),
					mustMake(t, OpGetGlobal, 2),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			t.Log(bytecode.Constants)
			t.Log(bytecode.Instructions)
			// assertBytecode(t, tt.wantBytecode, bytecode)
			vm := NewVM(bytecode)
			err := vm.Run()
			assert.NoError(t, err, "runtime error")
			got := vm.lastPoppedStackElem()
			assert.Equal(t, tt.wantStackTop, got)
		})
	}
}

func TestMapRange(t *testing.T) {
	tests := []testCase{
		{
			name: "map range",
			input: `x := ""
			for e := range {a:22 b:2}
				x = e
			end
			x = x
			`,
			wantStackTop: makeValue(t, 2),
			wantBytecode: &Bytecode{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
			t.Log(bytecode.Constants)
			t.Log(bytecode.Instructions)
			// assertBytecode(t, tt.wantBytecode, bytecode)
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
	case []any:
		return arrayVal{Elements: makeValues(t, v...)}
	case map[string]any:
		m := mapVal{
			order: make([]stringVal, 0),
			m:     make(map[stringVal]value),
		}
		for key, val := range v {
			m.m[stringVal(key)] = makeValue(t, val)
			m.order = append(m.order, stringVal(key))
		}
		return m
	case string:
		return stringVal(v)
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
