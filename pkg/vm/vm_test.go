package vm

import (
	"fmt"
	"testing"

	"evylang.dev/evy/pkg/compiler"
	"evylang.dev/evy/pkg/object"
	"evylang.dev/evy/pkg/parser"
)

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"one := 1\none = one", 1},
		{"two := 2\ntwo = two", 2},
		{"x := 1 + 2\nx = x", 3},
		{"x := 2 + 1\nx = x", 3},
		{"x := 2 - 1\nx = x", 1},
		{"x := 1 - 2\nx = x", -1},
		{"x := 2 - 3 + 2\nx = x", 1},
		{"x := 2 * 2\n x = x", 4},
		{"x := 2 * 1\n x = x", 2},
		{"x := 1 * 3 - 2 + 4\n x = x", 5},
		{"x := 2 / 2\n x = x", 1},
		{"x := 2 / 1\n x = x", 0.5},
		{"x := 4 / 2 - 2 + 4\n x = x", 4},
		{"x := 2 % 2\n x = x", 0},
		{"x := 1 % 2\n x = x", 1},
		{"x := 1 + 2 - 3 * 4 / 5 % 6\n x = x", 1},
		{"x := 2 + 2 / 2 \n x = x", 3},
		{"x := (2 + 2) / 2 \n x = x", 2},
	}

	runVmTests(t, tests)
}

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program, err := parser.Parse(tt.input, parser.Builtins{})
		if err != nil {
			t.Fatalf("parser error: %s", err)
		}

		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem, tt.input)
	}
}

func testExpectedObject(
	t *testing.T,
	expected interface{},
	actual object.Object,
	input string,
) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s. Input: %q", err, input)
		}
	}
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
	}

	return nil
}
