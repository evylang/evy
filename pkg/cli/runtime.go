// Package cli provides a runtime to for Evy CLI execution in terminal.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"evylang.dev/evy/pkg/evaluator"
)

// Runtime implements evaluator.Runtime.
type Runtime struct {
	evaluator.UnimplementedRuntime
	reader    *bufio.Reader
	SkipSleep bool
}

// NewRuntime returns an initialized cli runtime.
func NewRuntime() *Runtime {
	return &Runtime{reader: bufio.NewReader(os.Stdin)}
}

// Print prints s to stdout.
func (*Runtime) Print(s string) {
	fmt.Print(s)
}

// Cls clears the screen.
func (*Runtime) Cls() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("cannot clear screen", err)
	}
}

// Read reads a line of input from stdin and strips trailing newline.
func (rt *Runtime) Read() string {
	s, err := rt.reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return s[:len(s)-1] // strip trailing newline
}

// Sleep sleeps for dur. If the --skip-sleep flag is used, it does nothing.
func (rt *Runtime) Sleep(dur time.Duration) {
	if !rt.SkipSleep {
		time.Sleep(dur)
	}
}

// Yielder returns a no-op yielder for CLI evy as it is not needed. By
// contrast, browser Evy needs to explicitly hand over control to JS
// host with Yielder.
func (*Runtime) Yielder() evaluator.Yielder { return nil }
