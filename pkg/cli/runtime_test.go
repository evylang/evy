//go:build !tinygo

package cli

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
	"evylang.dev/evy/pkg/evaluator"
)

func TestGraphics(t *testing.T) {
	files, err := filepath.Glob("testdata/*.svg")
	assert.NoError(t, err)
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		t.Run(name, func(t *testing.T) {
			style := "border: 1px solid red; width: 400px; height: 400px"
			svgWriter := &bytes.Buffer{}
			rt := NewPlatform(WithSVG(style, "", ""), WithSkipSleep(true))

			eval := evaluator.NewEvaluator(rt)
			evyFilename := strings.TrimSuffix(file, filepath.Ext(file)) + ".evy"
			b, err := os.ReadFile(evyFilename)
			assert.NoError(t, err)
			err = eval.Run(string(b))
			assert.NoError(t, err)
			err = rt.WriteSVG(svgWriter)
			assert.NoError(t, err)

			b, err = os.ReadFile(file)
			assert.NoError(t, err)
			want := string(b)
			got := svgWriter.String()
			assert.Equal(t, want, got)
		})
	}
}

func TestPrintRead(t *testing.T) {
	readBuffer := &bytes.Buffer{}
	writeBuffer := &bytes.Buffer{}
	rt := &Platform{
		reader:           bufio.NewReader(readBuffer),
		writer:           writeBuffer,
		GraphicsPlatform: &evaluator.UnimplementedPlatform{},
	}
	readBuffer.WriteString("Hello world\n")
	s := rt.Read()
	assert.Equal(t, "Hello world", s)
	rt.Print("And good bye")
	want := "And good bye"
	got := writeBuffer.String()
	assert.Equal(t, want, got)
}
