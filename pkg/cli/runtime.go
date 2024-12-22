//go:build !tinygo

// Package cli provides an Evy platform to for Evy CLI execution in terminal.
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

// Platform implements evaluator.Platform.
type Platform struct {
	evaluator.GraphicsPlatform
	reader    *bufio.Reader
	writer    io.Writer
	clsFn     func()
	SkipSleep bool
}

// Option is used on Platform creation to set optional parameters.
type Option func(*Platform)

// WithSkipSleep sets the SkipSleep field Platform and is intended to be used
// with NewPlatform.
func WithSkipSleep(skipSleep bool) Option {
	return func(rt *Platform) {
		rt.SkipSleep = skipSleep
	}
}

// WithSVG sets up an SVG graphics platform and writes its output to the
// given writer.
func WithSVG(svgStyle string, svgWidth string, svgHeight string) Option {
	return func(rt *Platform) {
		svgRT := svg.NewGraphicsPlatform()
		svgRT.SVG.Style = svgStyle
		svgRT.SVG.Width = svgWidth
		svgRT.SVG.Height = svgHeight
		rt.GraphicsPlatform = svgRT
	}
}

// WithOutputWriter sets the text output writer, which defaults to os.Stdout.
func WithOutputWriter(w io.Writer) Option {
	return func(rt *Platform) {
		rt.writer = w
	}
}

// WithCls sets the action to be done for `cls` command.
func WithCls(clsFn func()) Option {
	return func(rt *Platform) {
		rt.clsFn = clsFn
	}
}

// NewPlatform returns an initialized cli platform.
func NewPlatform(options ...Option) *Platform {
	rt := &Platform{
		reader:           bufio.NewReader(os.Stdin),
		writer:           os.Stdout,
		GraphicsPlatform: &evaluator.UnimplementedPlatform{},
	}
	for _, opt := range options {
		opt(rt)
	}
	return rt
}

// Print prints s to stdout.
func (rt *Platform) Print(s string) {
	fmt.Fprint(rt.writer, s) //nolint:errcheck // no need to check for stdout
}

// Cls clears the screen.
func (rt *Platform) Cls() {
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
func (rt *Platform) Read() string {
	s, err := rt.reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return s[:len(s)-1] // strip trailing newline
}

// Sleep sleeps for dur. If the --skip-sleep flag is used, it does nothing.
func (rt *Platform) Sleep(dur time.Duration) {
	if !rt.SkipSleep {
		time.Sleep(dur)
	}
}

// Yielder returns a no-op yielder for CLI evy as it is not needed. By
// contrast, browser Evy needs to explicitly hand over control to JS
// host with Yielder.
func (*Platform) Yielder() evaluator.Yielder { return nil }

// WriteSVG writes the graphics output in SVG format to the writer set with
// option WithSVGWriter.
func (rt *Platform) WriteSVG(w io.Writer) error {
	graphicsRT := rt.GraphicsPlatform.(*svg.GraphicsPlatform)
	return graphicsRT.WriteSVG(w)
}
