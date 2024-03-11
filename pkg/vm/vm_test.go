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

func TestString(t *testing.T) {
	tests := []vmTestCase{
		{`x := "foobar"
x = x`, "foobar"},
		{`x := "foo" + "bar"
x = x`, "foobar"},
		{`x := "foobar" == "fubar"
x = x`, false},
		{`x := "foobar" != "fubar"
x = x`, true},
		{`x := "foobar" >= "fubar"
x = x`, false},
		{`x := "foobar" <= "fubar"
x = x`, true},
		{`x := "foobar" > "fubar"
x = x`, false},
		{`x := "foobar" < "fubar"
x = x`, true},
	}
	runVmTests(t, tests)
}

func TestArray(t *testing.T) {
	tests := []vmTestCase{
		{
			`x := [1 2 3]
x = x`, []any{1, 2, 3},
		},
		{
			`x := [1 "b" 3]
x = x`, []any{1, "b", 3},
		},
	}
	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{
			`x := 1
			if true
				x = 10
			end
			x = x`, 10,
		},
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
		{
			`x := 1
			if 1 > 2
				x = 10
			else if 1 < 2
				x = 20
			end
			x = x`, 20,
		},
		{
			`x := 1
			if 1 > 2
				x = 10
			else if 1 < 2
				x = 20
			else
				x = 100
			end
			x = x`, 20,
		},
		{
			`x := 1
			if x > 2 
				x = 10
			else if x < 2
				x = 20
			else
				x = 100
			end
			x = x`, 20,
		},
		{
			`x := 3
			if x > 2 
				x = 10
			else if x == 10
				x = 0
			else
				x = 100
			end
			x = x`, 10,
		},
		{
			`x := 1
			if x == 1
				x = 2
			else if x == 2
				x = 3
			else if x == 3
				x = 4
			else 
				x = 5
			end	
			x = x`, 2,
		},
	}

	runVmTests(t, tests)
}

type vmTestCase struct {
	input    string
	expected any
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
	expected any,
	actual object.Object,
	input string,
) {
	t.Helper()

	switch expected := expected.(type) {
	case bool:
		if err := testBooleanObject(expected, actual); err != nil {
			t.Errorf("testBooleanObject failed: %s. Input: %q", err, input)
		}
	case int:
		if err := testIntegerObject(float64(expected), actual); err != nil {
			t.Errorf("testIntegerObject failed: %s. Input: %q", err, input)
		}
	case float64:
		if err := testIntegerObject(expected, actual); err != nil {
			t.Errorf("testIntegerObject failed: %s. Input: %q", err, input)
		}
	case string:
		if err := testStringObject(expected, actual); err != nil {
			t.Errorf("testStringObject failed: %s. Input: %q", err, input)
		}
	case []any:
		if err := testArrayObject(t, expected, actual); err != nil {
			t.Errorf("testStringObject failed: %s. Input: %q", err, input)
		}
	default:
		t.Errorf("type of expected (%T) not handled", expected)
	}
}

func testArrayObject(t *testing.T, expected []any, actual object.Object) error {
	arr, ok := actual.(*object.Array)
	if !ok {
		return fmt.Errorf("object is not Array. got=%T (%+v)",
			actual, actual)
	}
	for i := 0; i < len(expected); i++ {
		testExpectedObject(t, expected[i], arr.Elements[i], "")
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
