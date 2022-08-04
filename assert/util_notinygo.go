//go:build !tinygo

package assert

import (
	"testing"
)

func fatalf(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Fatalf(format, args...)
}
