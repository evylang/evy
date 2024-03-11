package compiler

import (
	"fmt"
	"testing"

	"evylang.dev/evy/pkg/code"
	"evylang.dev/evy/pkg/object"
	"evylang.dev/evy/pkg/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := 1 + 2
x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 - 2
x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSubtract),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 - 2 + 3
x = x`,
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSubtract),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 + 2 - 3
x = x`,
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSubtract),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 * 2 * 3
x = x`,
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMultiply),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpMultiply),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		}, {
			input: `x := 1 * 2 + 3 - 4
x = x`,
			expectedConstants: []interface{}{1, 2, 3, 4},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMultiply),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSubtract),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 - 2 + 3 * 4
x = x`,
			expectedConstants: []interface{}{1, 2, 3, 4},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSubtract),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMultiply),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			// x=((2+((3*3)*3))+2)
			input: `x := 2 + 3 * 3 * 3 + 2
x = x`,
			expectedConstants: []interface{}{2, 3, 3, 3, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0), // 2
				code.Make(code.OpConstant, 1), // (3
				code.Make(code.OpConstant, 2), // 3)
				code.Make(code.OpMultiply),    // (3 * 3)
				code.Make(code.OpConstant, 3), // 3))
				code.Make(code.OpMultiply),    // ((3 * 3) * 3)
				code.Make(code.OpAdd),         // (2 + ((3 * 3) * 3))
				code.Make(code.OpConstant, 4), // 2
				code.Make(code.OpAdd),         // ((2 + ((3 * 3) * 3)) + 2)
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 / 2 / 3
x = x`,
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDivide),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpDivide),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 % 2
x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpModulo),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 2 + 2 / 2
x = x`,
			expectedConstants: []interface{}{2, 2, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpDivide),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := (2 + 2) / 2
x = x`,
			expectedConstants: []interface{}{2, 2, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpDivide),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 + 2 - 3 * 4 / 5 % 6
x = x`,
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMultiply),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpDivide),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpModulo),
				code.Make(code.OpSubtract),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 + -1
x = x`,
			expectedConstants: []interface{}{1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMinus),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBool(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := true
			x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `two := false
			two = two`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},

		{
			input: `x := 1 > 2
			x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 >= 2
			x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThanEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 < 2
			x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpLessThan),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 <= 2
			x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpLessThanEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 == 2
			x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 1 != 2
			x = x`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := true == true
			x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpTrue),
				code.Make(code.OpEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := true != false
			x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := !true
			x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := !!true
			x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpBang),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestString(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := "string"
x = x`,
			expectedConstants: []interface{}{"string"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := "string" + "concat"
x = x`,
			expectedConstants: []interface{}{"string", "concat"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := "string" == "concat"
x = x`,
			expectedConstants: []interface{}{"string", "concat"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := "string" != "concat"
x = x`,
			expectedConstants: []interface{}{"string", "concat"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		}, {
			input: `x := "a" >= "b"
x = x`,
			expectedConstants: []interface{}{"a", "b"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThanEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		}, {
			input: `x := "a" > "b"
x = x`,
			expectedConstants: []interface{}{"a", "b"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		}, {
			input: `x := "a" <= "b"
x = x`,
			expectedConstants: []interface{}{"a", "b"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpLessThanEqual),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		}, {
			input: `x := "a" < "b"
x = x`,
			expectedConstants: []interface{}{"a", "b"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpLessThan),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := "hello"
			y := x[4]
			y = y`,
			expectedConstants: []interface{}{"hello", 4},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpIndex),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := "abc"
			y := x[-1]
			y = y`,
			expectedConstants: []interface{}{"abc", 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMinus),
				code.Make(code.OpIndex),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := "abc"
			y := x[0:1]
			y = y`,
			expectedConstants: []interface{}{"abc", 0, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := "abc"
			y := x[:1]
			y = y`,
			expectedConstants: []interface{}{"abc", 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpNull),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := "abc"
			y := x[1:]
			y = y`,
			expectedConstants: []interface{}{"abc", 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNull),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := "abc"
			y := x[:]
			y = y`,
			expectedConstants: []interface{}{"abc"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpNull),
				code.Make(code.OpNull),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArray(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := []
		x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := [1 2 3]
		x = x`,
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := ["1" 2 "3"]
		x = x`,
			expectedConstants: []any{"1", 2, "3"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := ["1" 2 "3"]
			y := x[0]
			y = y`,
			expectedConstants: []any{"1", 2, "3", 0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpIndex),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := [1 2 3]
			y := x[0:1]
			y = y`,
			expectedConstants: []any{1, 2, 3, 0, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := [1 2 3]
			y := x[:1]
			y = y`,
			expectedConstants: []any{1, 2, 3, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpNull),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := [1 2 3]
			y := x[1:]
			y = y`,
			expectedConstants: []any{1, 2, 3, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpNull),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `x := [1 2 3]
			y := x[:]
			y = y`,
			expectedConstants: []any{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpNull),
				code.Make(code.OpNull),
				code.Make(code.OpSlice),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestMap(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := {}
		x = x`,
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpMap, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := {a: 1 b: 2}
		x = x`,
			expectedConstants: []interface{}{"a", 1, "b", 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMap, 4),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := {a: 1 b: 2}
		y := x["a"]
		y = y`,
			expectedConstants: []any{"a", 1, "b", 2, "a"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMap, 4),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpIndex),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := 1
			if x == 1
				x = 2
			end
			x = x`,
			expectedConstants: []interface{}{1, 1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpJumpNotTruthy, 25),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpJump, 25),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 10
			if x > 5
				x = 20
			else
				x = 5
			end
			x = x`,
			expectedConstants: []interface{}{10, 5, 20, 5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpJumpNotTruthy, 25),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpJump, 31),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `x := 10
			if x > 10
				x = 10
			else if x > 5
				x = 5
			else
				x = 1
			end
			x = x`,
			expectedConstants: []interface{}{10, 10, 10, 5, 5, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpJumpNotTruthy, 25),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpJump, 50), // If the first condition is true, jump to the end after statement
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpJumpNotTruthy, 41),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpJump, 50),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalVarStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `x := 1
			x = x`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `two := 2
			two = two`,
			expectedConstants: []interface{}{2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `three := 3
			three = 30
			three = three`,
			expectedConstants: []interface{}{3, 30},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `one := 1
			two := one
			three := two
			three = three`,
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpSetGlobal, 2),
				code.Make(code.OpGetGlobal, 2),
				code.Make(code.OpSetGlobal, 2),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		program, err := parser.Parse(tt.input, parser.Builtins{})
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}
		t.Log(program.String())

		compiler := New()
		if err := compiler.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()
		testInstructions(t, tt.expectedInstructions, bytecode.Instructions, tt.input)

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(
	t *testing.T,
	expected []code.Instructions,
	actual code.Instructions,
	input string,
) {
	concatted := concatInstructions(expected)
	if len(actual) != len(concatted) {
		t.Fatalf("input %s\nwrong instructions length.\nwant=%s\ngot =%s", input, concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			t.Fatalf("input %s\nwrong instruction at %d.\nwant=%s\ngot =%s", input, i, concatted, actual)
		}
	}
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}

	return out
}

func testConstants(
	t *testing.T,
	expected []any,
	actual []object.Object,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			if err := testIntegerObject(float64(constant), actual[i]); err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case float64:
			if err := testIntegerObject(constant, actual[i]); err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case string:
			if err := testStringObject(constant, actual[i]); err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s", i, err)
			}
		default:
			t.Errorf("unexpected type %T", constant)
		}
	}
	return nil
}

func testIntegerObject(expected float64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%f, want=%f", result.Value, expected)
	}

	return nil
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%s, want=%s",
			result.Value, expected)
	}

	return nil
}
