//go:build !tinygo

// Package cli provides an Evy runtime to for Evy CLI execution in terminal.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"evylang.dev/evy/pkg/cli/svg"
	"evylang.dev/evy/pkg/evaluator"
)

// Runtime implements evaluator.Runtime.
type Runtime struct {
	evaluator.GraphicsRuntime
	reader    *bufio.Reader
	writer    io.Writer
	clsFn     func()
	SkipSleep bool
}

// Option is used on Runtime creation to set optional parameters.
type Option func(*Runtime)

// WithSkipSleep sets the SkipSleep field Runtime and is intended to be used
// with NewRuntime.
func WithSkipSleep(skipSleep bool) Option {
	return func(rt *Runtime) {
		rt.SkipSleep = skipSleep
	}
}

// WithSVG sets up an SVG graphics runtime and writes its output to the
// given writer.
func WithSVG(svgStyle string, svgWidth string, svgHeight string) Option {
	return func(rt *Runtime) {
		svgRT := svg.NewGraphicsRuntime()
		svgRT.SVG.Style = svgStyle
		svgRT.SVG.Width = svgWidth
		svgRT.SVG.Height = svgHeight
		rt.GraphicsRuntime = svgRT
	}
}

// WithOutputWriter sets the text output writer, which defaults to os.Stdout.
func WithOutputWriter(w io.Writer) Option {
	return func(rt *Runtime) {
		rt.writer = w
	}
}

// WithCls sets the action to be done for `cls` command.
func WithCls(clsFn func()) Option {
	return func(rt *Runtime) {
		rt.clsFn = clsFn
	}
}

// NewRuntime returns an initialized cli runtime.
func NewRuntime(options ...Option) *Runtime {
	rt := &Runtime{
		reader:          bufio.NewReader(os.Stdin),
		writer:          os.Stdout,
		GraphicsRuntime: &evaluator.UnimplementedRuntime{},
	}
	for _, opt := range options {
		opt(rt)
	}
	return rt
}

// Print prints s to stdout.
func (rt *Runtime) Print(s string) {
	fmt.Fprint(rt.writer, s) //nolint:errcheck // no need to check for stdout
}

// Cls clears the screen.
func (rt *Runtime) Cls() {
	if rt.clsFn != nil {
		rt.clsFn()
		return
	}
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

// WriteSVG writes the graphics output in SVG format to the writer set with
// option WithSVGWriter.
func (rt *Runtime) WriteSVG(w io.Writer) error {
	graphicsRT := rt.GraphicsRuntime.(*svg.GraphicsRuntime)
	return graphicsRT.WriteSVG(w)
}
