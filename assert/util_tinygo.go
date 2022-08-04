//go:build tinygo

package assert

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

var (
	exit           = os.Exit
	out  io.Writer = os.Stdout
)

func fatalf(t *testing.T, format string, args ...any) {
	t.Helper()
	fmt.Fprintln(out, "--- FAIL:", t.Name())
	s := fmt.Sprintf(format+"\n", args...)
	fmt.Fprint(out, decorate(s))
	fmt.Fprintln(out, "FAIL")
	exit(1)
}

func decorate(s string) string {
	buf := new(strings.Builder)
	// Every line is indented at least 4 spaces.
	buf.WriteString("    ")
	lines := strings.Split(s, "\n")
	if l := len(lines); l > 1 && lines[l-1] == "" {
		lines = lines[:l-1]
	}
	for i, line := range lines {
		if i > 0 {
			// Second and subsequent lines are indented an additional 4 spaces.
			buf.WriteString("\n        ")
		}
		buf.WriteString(line)
	}
	buf.WriteByte('\n')
	return buf.String()
}
