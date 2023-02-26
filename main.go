//go:build !tinygo

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"foxygo.at/evy/pkg/evaluator"
	"foxygo.at/evy/pkg/lexer"
	"foxygo.at/evy/pkg/parser"
	"github.com/alecthomas/kong"
)

var version string = "v0.0.0"

const description = `
evy is a tool for managing evy source code.
`

type config struct {
	Version  kong.VersionFlag `short:"V" help:"Print version information"`
	Run      runCmd           `cmd:"" help:"Run evy program"`
	Tokenize tokenizeCmd      `cmd:"" help:"Tokenize evy program"`
	Parse    parseCmd         `cmd:"" help:"Parse evy program"`
}

type runCmd struct {
	Source string `arg:"" help:"Source file. Default stdin" default:"-"`
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
	eval.Run(string(b))
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
	builtinDecls := evaluator.DefaulParserBuiltins(newRuntime())
	result := parser.Run(string(b), builtinDecls)
	fmt.Println(result)
	return nil
}

func main() {
	kctx := kong.Parse(&config{},
		kong.Description(description),
		kong.Vars{"version": version},
	)
	kctx.FatalIfErrorf(kctx.Run())
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
