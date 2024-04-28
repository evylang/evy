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
}

func TestVMGlobals(t *testing.T) {
	tests := []testCase{
		{
			name:         "inferred declaration",
			input:        "x := 1",
			wantStackTop: makeValue(t, 1),
		},
		{
			name: "assignment",
			input: `x := 1
			y := x
			y = x + 1
			y = y`,
			wantStackTop: makeValue(t, 2),
		},
		{
			name: "index assignment",
			input: `x := [1 2 3]
			x[0] = x[2]
			x[2] = 1
			x = x`,
			wantStackTop: makeValue(t, []any{3, 2, 1}),
		},
		{
			name: "nested index assignment",
			input: `x := [[1 2] [3 4]]
			x[0][0] = x[0][1]
			x = x`,
			wantStackTop: makeValue(t, []any{[]any{2, 2}, []any{3, 4}}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
		{
			name:         "literal false",
			input:        "x := false",
			wantStackTop: makeValue(t, false),
		},
		{
			name:         "not operator",
			input:        "x := !true",
			wantStackTop: makeValue(t, false),
		},
		{
			name:         "equal operator",
			input:        "x := 1 == 1",
			wantStackTop: makeValue(t, true),
		},
		{
			name:         "not operator",
			input:        "x := 1 != 1",
			wantStackTop: makeValue(t, false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
		{
			name:         "subtraction",
			input:        "x := 2 - 1",
			wantStackTop: makeValue(t, 1),
		},
		{
			name:         "multiplication",
			input:        "x := 2 * 1",
			wantStackTop: makeValue(t, 2),
		},
		{
			name:         "division",
			input:        "x := 2 / 1",
			wantStackTop: makeValue(t, 2),
		},
		{
			name:         "modulo",
			input:        "x := 2 % 1",
			wantStackTop: makeValue(t, 0),
		},
		{
			name:         "float modulo",
			input:        "x := 2.5 % 1.3",
			wantStackTop: makeValue(t, 1.2),
		},
		{
			name:         "minus operator",
			input:        "x := -1",
			wantStackTop: makeValue(t, -1),
		},
		{
			name:         "all operators",
			input:        "x := 1 + 2 - 3 * 4 / 5 % 6",
			wantStackTop: makeValue(t, 1+2-math.Mod(3.0*4.0/5.0, 6.0)),
		},
		{
			name:         "grouped expressions",
			input:        "x := (1 + 2 - 3) * 4 / 5 % 6",
			wantStackTop: makeValue(t, (1+2-3)*4/math.Mod(5.0, 6.0)),
		},
		{
			name:         "less than",
			input:        "x := 1 < 2",
			wantStackTop: makeValue(t, true),
		},
		{
			name:         "less than or equal",
			input:        "x := 1 <= 2",
			wantStackTop: makeValue(t, true),
		},
		{
			name:         "greater than",
			input:        "x := 1 > 2",
			wantStackTop: makeValue(t, false),
		},
		{
			name:         "greater than or equal",
			input:        "x := 1 >= 2",
			wantStackTop: makeValue(t, false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
		{
			name:         "concatenate",
			input:        `x := "a" + "b"`,
			wantStackTop: makeValue(t, "ab"),
		},
		{
			name:         "less than",
			input:        `x := "a" < "b"`,
			wantStackTop: makeValue(t, true),
		},
		{
			name:         "less than or equal",
			input:        `x := "a" <= "b"`,
			wantStackTop: makeValue(t, true),
		},
		{
			name:         "greater than",
			input:        `x := "a" > "b"`,
			wantStackTop: makeValue(t, false),
		},
		{
			name:         "greater than or equal",
			input:        `x := "a" >= "b"`,
			wantStackTop: makeValue(t, false),
		},
		{
			name:         "index",
			input:        `x := "abc"[0]`,
			wantStackTop: makeValue(t, "a"),
		},
		{
			name:         "negative index",
			input:        `x := "abc"[-1]`,
			wantStackTop: makeValue(t, "c"),
		},
		{
			name:         "slice",
			input:        `x := "abc"[1:3]`,
			wantStackTop: makeValue(t, "bc"),
		},
		{
			name:         "negative slice",
			input:        `x := "abc"[-3:-1]`,
			wantStackTop: makeValue(t, "ab"),
		},
		{
			name:         "slice default start",
			input:        `x := "abc"[:1]`,
			wantStackTop: makeValue(t, "a"),
		},
		{
			name:         "slice default end",
			input:        `x := "abc"[1:]`,
			wantStackTop: makeValue(t, "bc"),
		},
		{
			name:         "slice start equals length",
			input:        `x := "abc"[3:]`,
			wantStackTop: makeValue(t, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
		{
			name:         "assignment",
			input:        "x := [1 2 3]",
			wantStackTop: makeValue(t, []any{1, 2, 3}),
		},
		{
			name:         "concatenate",
			input:        "x := [1 2] + [3 4]",
			wantStackTop: makeValue(t, []any{1, 2, 3, 4}),
		},
		{
			name:         "repetition",
			input:        `x := [1 2] * 3`,
			wantStackTop: makeValue(t, []any{1, 2, 1, 2, 1, 2}),
		},
		{
			name:         "zero repetition",
			input:        `x := [1 2] * 0`,
			wantStackTop: makeValue(t, []any{}),
		},
		{
			name: "concatenate preserve original",
			input: `x := [1 2]
			y := [3 4]
			y = x + y
			x = x`,
			wantStackTop: makeValue(t, []any{1, 2}),
		},
		{
			name:         "index",
			input:        `x := [1 2 3][0]`,
			wantStackTop: makeValue(t, 1),
		},
		{
			name:         "negative index",
			input:        `x := [1 2 3][-1]`,
			wantStackTop: makeValue(t, 3),
		},
		{
			name:         "slice",
			input:        `x := [1 2 3][1:3]`,
			wantStackTop: makeValue(t, []any{2, 3}),
		},
		{
			name:         "negative slice",
			input:        `x := [1 2 3][-3:-1]`,
			wantStackTop: makeValue(t, []any{1, 2}),
		},
		{
			name:         "slice default start",
			input:        `x := [1 2 3][:1]`,
			wantStackTop: makeValue(t, []any{1}),
		},
		{
			name:         "slice default end",
			input:        `x := [1 2 3][1:]`,
			wantStackTop: makeValue(t, []any{2, 3}),
		},
		{
			name:         "slice start equals length",
			input:        `x := [1 2 3][3:]`,
			wantStackTop: makeValue(t, []any{}),
		},
		{
			name: "slice preserve original",
			input: `x := [1 2 3]
			y := x[1:]
			y[0] = 8
			x = x`,
			wantStackTop: makeValue(t, []any{1, 2, 3}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
		{
			name:         "assignment",
			input:        "x := {a: 1 b: 2}",
			wantStackTop: makeValue(t, map[string]any{"a": 1, "b": 2}),
		},
		{
			name:         "index",
			input:        `x := {a: 1 b: 2}["b"]`,
			wantStackTop: makeValue(t, 2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
		},
		{
			name: "skip loop",
			input: `x := 0
			while x > 5
				x = x + 1
			end
			x = x`,
			wantStackTop: makeValue(t, 0),
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := compileBytecode(t, tt.input)
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
