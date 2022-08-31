package assert

import (
	"reflect"
	"testing"
)

func NoError(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Helper()
	fatalf(t, "err: %v", err)
}

func Equal(t *testing.T, want, got any) {
	if equal(want, got) {
		return
	}
	t.Helper()
	fatalf(t, "want != got\n%#v\n%#v", want, got)
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
