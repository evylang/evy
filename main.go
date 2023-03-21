//go:build !tinygo

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"foxygo.at/evy/pkg/evaluator"
	"foxygo.at/evy/pkg/lexer"
	"foxygo.at/evy/pkg/parser"
	"github.com/alecthomas/kong"
)

var (
	version         string = "v0.0.0"
	errBadWriteFlag        = errors.New("cannot use -w without files")
	errNotFormatted        = errors.New("not formatted")
	errParse               = errors.New("parse error")
)

const description = `
evy is a tool for managing evy source code.
`

type config struct {
	Version kong.VersionFlag `short:"V" help:"Print version information"`
	Run     runCmd           `cmd:"" help:"Run evy program"`
	Fmt     fmtCmd           `cmd:"" help:"Fmt evy files"`

	Tokenize tokenizeCmd `cmd:"" help:"Tokenize evy program" hidden:""`
	Parse    parseCmd    `cmd:"" help:"Parse evy program" hidden:""`
}

func main() {
	kopts := []kong.Option{
		kong.Description(description),
		kong.Vars{"version": version},
	}
	kctx := kong.Parse(&config{}, kopts...)
	kctx.FatalIfErrorf(kctx.Run())
}

type runCmd struct {
	Source string `arg:"" help:"Source file. Default stdin" default:"-"`
}

type fmtCmd struct {
	Write bool     `short:"w" help:"update .evy file" xor:"mode"`
	Check bool     `short:"c" help:"check if already formatted" xor:"mode"`
	Files []string `arg:"" optional:"" help:"Source files. Default stdin"`
}

type tokenizeCmd struct {
	Source string `arg:"" help:"Source file. Default stdin" default:"-"`
}

type parseCmd struct {
	Source string `arg:"" help:"Source file. Default stdin" default:"-"`
}

func (c *runCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	builtins := evaluator.DefaultBuiltins(newRuntime())
	eval := evaluator.NewEvaluator(builtins)
	return eval.Run(string(b))
}

func (c *fmtCmd) Run() error {
	if len(c.Files) == 0 {
		if c.Write {
			return errBadWriteFlag
		}
		return format(os.Stdin, os.Stdout, c.Check)
	}
	var out io.StringWriter = os.Stdout
	for _, filename := range c.Files {
		in, err := os.Open(filename)
		if err != nil {
			return err
		}
		if c.Write {
			out, err = os.CreateTemp("", "evy")
			if err != nil {
				return fmt.Errorf("%s: %w", filename, err)
			}
		}
		if err := format(in, out, c.Check); err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		if err := in.Close(); err != nil {
			return err
		}
		if c.Write {
			tempFile := out.(*os.File)
			if err := tempFile.Close(); err != nil {
				return fmt.Errorf("%s: %w", filename, err)
			}
			if err := os.Rename(tempFile.Name(), filename); err != nil {
				return fmt.Errorf("%s: %w", filename, err)
			}
		}
	}
	return nil
}

func format(r io.Reader, w io.StringWriter, checkOnly bool) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	in := string(b)
	parserBuiltins := evaluator.DefaultBuiltins(newRuntime()).ParserBuiltins()
	prog, err := parser.Parse(in, parserBuiltins)
	if err != nil {
		return fmt.Errorf("%w: %w", errParse, parser.TruncateError(err, 8))
	}
	out := prog.Format()
	if checkOnly {
		if in != out {
			return errNotFormatted
		}
		return nil
	}
	if _, err := w.WriteString(out); err != nil {
		return err
	}
	return nil
}

func (c *tokenizeCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	result := lexer.Run(string(b))
	fmt.Println(result)
	return nil
}

func (c *parseCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	builtinDecls := evaluator.DefaultBuiltins(newRuntime()).ParserBuiltins()
	ast, err := parser.Parse(string(b), builtinDecls)
	if err != nil {
		return fmt.Errorf("%w: %w", errParse, parser.TruncateError(err, 8))
	}
	fmt.Println(ast.String())
	return nil
}

func fileBytes(filename string) ([]byte, error) {
	if filename == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(filename)
}

func newRuntime() *evaluator.Runtime {
	reader := bufio.NewReader(os.Stdin)
	return &evaluator.Runtime{
		Print: func(s string) { fmt.Print(s) },
		Read:  func() string { s, _ := reader.ReadString('\n'); return s },
		Sleep: time.Sleep,
	}
}
