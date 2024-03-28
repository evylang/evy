package bytecode

import (
	"errors"
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
			assert.Equal(t, true, errors.Is(err, ErrBounds))
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
	assert.Equal(t, true, errors.Is(err, ErrMapKey))
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
		m := mapVal{}
		for key, val := range v {
			m[key] = makeValue(t, val)
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
