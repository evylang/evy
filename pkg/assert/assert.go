// Package assert is a minimal testing utility package with inline
// assertions.
package assert

import (
	"fmt"
	"reflect"
	"testing"
)

// NoError is a testing utility function that immediately fails the test
// if err is not nil and prints the optional message with arguments in
// fmt.Printf syntax.
func NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatalf("err: %v%s", err, format(msgAndArgs...))
}

// Equal is a testing utility function that immediately fails the test
// if want is not equal to got and prints the optional message with
// arguments in fmt.Printf syntax.
func Equal(t *testing.T, want, got any, msgAndArgs ...interface{}) {
	t.Helper()
	if equal(want, got) {
		return
	}
	t.Fatalf("want != got\n%#v\n%#v%s", want, got, format(msgAndArgs...))
}

func equal(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	return aValue == bValue
}

func format(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	return fmt.Sprintf("\n"+msgAndArgs[0].(string), msgAndArgs[1:]...)
}
