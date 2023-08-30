//go:build !tinygo

// Evy compiles, runs, and formats Evy source code.
//
// Evy on the command line supports all [built-in functions] except for
// graphics functions and event handlers. Only the web interface on
// [evy.dev/play] supports graphics and events.
//
// The Evy toolchain has two subcommands: run and fmt.
//
//	Usage: evy <command>
//
//	evy is a tool for managing evy source code.
//
//	Flags:
//	  -h, --help       Show context-sensitive help.
//	  -V, --version    Print version information
//
//	Commands:
//	  run [<source>]
//	    Run Evy program.
//
//	  fmt [<files> ...]
//	    Format Evy files.
//
//	Run "evy <command> --help" for more information on a command.
//
// [built-in functions]: https://github.com/foxygoat/evy/blob/master/docs/builtins.md
// [evy.dev/play]: https://evy.dev/play
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
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

// cliRuntime implements evaluator.Runtime.
type cliRuntime struct {
	evaluator.UnimplementedRuntime
	reader    *bufio.Reader
	skipSleep bool
}

func newCLIRuntime() *cliRuntime {
	return &cliRuntime{reader: bufio.NewReader(os.Stdin)}
}

// Print prints s to stdout.
func (*cliRuntime) Print(s string) {
	fmt.Print(s)
}

// Cls clears the screen.
func (*cliRuntime) Cls() {
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
func (rt *cliRuntime) Read() string {
	s, err := rt.reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return s[:len(s)-1] // strip trailing newline
}

// Sleep sleeps for dur. If the --skip-sleep flag is used, it does nothing.
func (rt *cliRuntime) Sleep(dur time.Duration) {
	if !rt.skipSleep {
		time.Sleep(dur)
	}
}

// Yielder returns a no-op yielder for CLI evy as it is not needed. By
// contrast, browser Evy needs to explicitly hand over control to JS
// host with Yielder.
func (*cliRuntime) Yielder() evaluator.Yielder { return nil }

const description = `
evy is a tool for managing evy source code.
`

type config struct {
	Version kong.VersionFlag `short:"V" help:"Print version information"`
	Run     runCmd           `cmd:"" help:"Run Evy program."`
	Fmt     fmtCmd           `cmd:"" help:"Format Evy files."`

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
	Source    string `arg:"" help:"Source file. Default stdin" default:"-"`
	SkipSleep bool   `help:"skip evy sleep command" env:"EVY_SKIP_SLEEP"`
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

// Run implements the `evy run` CLI command, called by the Kong API.
func (c *runCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	rt := newCLIRuntime()
	rt.skipSleep = c.SkipSleep
	builtins := evaluator.DefaultBuiltins(rt)
	eval := evaluator.NewEvaluator(builtins)
	err = eval.Run(string(b))
	handlEvyErr(err)
	return nil
}

func handlEvyErr(err error) {
	if err == nil {
		return
	}
	var exitErr evaluator.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(int(exitErr))
	}
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

// Run implements the `evy fmt` CLI command, called by the Kong API.
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
	parserBuiltins := evaluator.DefaultBuiltins(newCLIRuntime()).ParserBuiltins()
	prog, err := parser.Parse(in, parserBuiltins)
	if err != nil {
		return fmt.Errorf("%w: %w", errParse, truncateError(err))
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

// Run implements the hidden `evy tokenize` CLI command, called by the
// Kong API.
func (c *tokenizeCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	l := lexer.New(string(b))
	tok := l.Next()
	for ; tok.Type != lexer.EOF; tok = l.Next() {
		fmt.Println(tok)
	}
	fmt.Println(tok)
	fmt.Println()
	return nil
}

// Run implements the hidden `evy parse` CLI command, called by the
// Kong API.
func (c *parseCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	builtinDecls := evaluator.DefaultBuiltins(newCLIRuntime()).ParserBuiltins()
	ast, err := parser.Parse(string(b), builtinDecls)
	if err != nil {
		return fmt.Errorf("%w: %w", errParse, truncateError(err))
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

func truncateError(err error) error {
	var parseErrors parser.Errors
	if errors.As(err, &parseErrors) {
		return parseErrors.Truncate(8)
	}
	return err
}
