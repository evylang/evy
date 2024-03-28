// Package assert is a minimal testing utility package with inline
// assertions.
package assert

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// Error immediately fails a test function if err does not contain
// the target error in its tree.
func Error(t *testing.T, want, got error) {
	t.Helper()
	if errors.Is(got, want) {
		return
	}
	t.Fatalf("want != got\n%v\n%v", want, got)
}

// NoError immediately fails a test function if err is not nil. It
// prints an optional formatted message with arguments.
func NoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		return
	}
	t.Fatalf("err: %v%s", err, format(msgAndArgs...))
}

// Equal immediately fails a test function if want is not equal to
// got. It prints an optional formatted message with arguments.
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
