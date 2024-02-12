package vm

import (
	"fmt"
	"math"
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
		{"x := 1 / 2\n x = x", 0.5},
		{"x := 4 / 2 - 2 + 4\n x = x", 4},
		{"x := 2 % 2\n x = x", 0},
		{"x := 1 % 2\n x = x", 1},
		{"x := 1 + 2 - 3 * 4 / 5 % 6\n x = x", 1 + 2 - math.Mod(3.0*4.0/5.0, 6.0)},
		{"x := 2 + 2 / 2 \n x = x", 3},
		{"x := (2 + 2) / 2 \n x = x", 2},
		{"x := (5 + 10 * 2 + 15 / 3) * 2 + -10 \n x = x", 50},
	}

	runVmTests(t, tests)
}

func TestBool(t *testing.T) {
	tests := []vmTestCase{
		{"x := true\nx = x", true},
		{"x := false\nx = x", false},
		{"x := 1 < 2\nx = x", true},
		{"x := 1 > 2\nx = x", false},
		{"x := 1 < 1\nx = x", false},
		{"x := 1 > 1\nx = x", false},
		{"x := 1 <= 1\nx = x", true},
		{"x := 2 <= 1\nx = x", false},
		{"x := 1 >= 1\nx = x", true},
		{"x := 1 >= 2\nx = x", false},
		{"x := 1 == 1\nx = x", true},
		{"x := 1 != 1\nx = x", false},
		{"x := 1 == 2\nx = x", false},
		{"x := 1 != 2\nx = x", true},
		{"x := true == true\nx = x", true},
		{"x := false == false\nx = x", true},
		{"x := true == false\nx = x", false},
		{"x := true != false\nx = x", true},
		{"x := false != true\nx = x", true},
		{"x := (1 < 2) == true\nx = x", true},
		{"x := (1 < 2) == false\nx = x", false},
		{"x := (1 > 2) == true\nx = x", false},
		{"x := (1 > 2) == false\nx = x", true},
		{"x := !!true\nx = x", true},
		{"x := !!!!!!!!true\nx = x", true},
		{"x := true\n x = !!!!x", true},
		{"x := true\n x = x != !x\n", true},
	}

	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		// {
		// 	`x := 1
		// 	if true
		// 		x = 10
		// 	end
		// 	x = x`, 10,
		// },
		{
			`x := 1
			if true
				x = 10
			else 
				x = 20 
			end
			x = x`, 10,
		},
		{
			`x := 1
			if false
				x = 10
			else 
				x = 20
			end
			x = x`, 20,
		},
		{
			`x := 1
			if 1 < 2
				x = 10
			end
			x = x`, 10,
		},
		{
			`x := 1
			if 1 < 2
				x = 10  
			else 
				x = 20
			end
			x = x`, 10,
		},
		{
			`x := 1
			if 1 > 2 
				x = 10 
			else 
				x = 20 
			end
			x = x`, 20,
		},
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
		// TODO: remove when done developing
		t.Log(program.String())

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
	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s. Input: %q", err, input)
		}
	case int:
		err := testIntegerObject(float64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s. Input: %q", err, input)
		}
	case float64:
		err := testIntegerObject(expected, actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s. Input: %q", err, input)
		}
	default:
		t.Errorf("type of expected (%T) not handled", expected)
	}
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
	}
	return nil
}

func testIntegerObject(expected float64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%f, want=%f",
			result.Value, expected)
	}

	return nil
}
