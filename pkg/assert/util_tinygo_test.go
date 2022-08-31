//go:build tinygo

package assert

import (
	"bytes"
	"testing"
)

func TestFatalf(t *testing.T) {
	savedExit, savedOut := exit, out
	defer func() { exit, out = savedExit, savedOut }()

	exitCode := 0
	exit = func(code int) { exitCode = code }
	buf := bytes.Buffer{}
	out = &buf

	fatalf(t, "%s: %d\n(by deep thought)", "answer", 42)
	got := buf.String()
	want := `--- FAIL: TestFatalf
    answer: 42
        (by deep thought)
FAIL
`
	if want != got {
		t.Errorf("want != got\n%v\n%v", want, got)
	}
	if exitCode != 1 {
		t.Errorf("exitCode expected: 1, got: %d", exitCode)
	}
}
