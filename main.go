//go:build !tinygo

// Evy parses, runs, and formats Evy source code.
//
// Evy on the command line supports all [built-in functions] except for
// graphics functions and event handlers. Only the web interface on
// [play.evy.dev] supports graphics and events.
//
// The Evy toolchain has two subcommands: run and fmt.
//
//	Usage: evy <command> [flags]
//
//	evy is a tool for managing evy source code.
//
//	Flags:
//	  -h, --help       Show context-sensitive help.
//	  -V, --version    Print version information
//
//	Commands:
//	  run [<source>] [flags]
//	    Run Evy program.
//
//	  fmt [<files> ...] [flags]
//	    Format Evy files.
//
//	  serve export <dir> [flags]
//	    Export embedded content.
//
//	  serve start [flags]
//	    Start web server, default for "evy serve".
//
//	Run "evy <command> --help" for more information on a command.
//
// [built-in functions]: https://github.com/evylang/evy/blob/main/docs/builtins.md
// [play.evy.dev]: https://play.evy.dev
package main

import (
	"cmp"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"evylang.dev/evy/pkg/cli"
	"evylang.dev/evy/pkg/evaluator"
	"evylang.dev/evy/pkg/lexer"
	"evylang.dev/evy/pkg/parser"
	"github.com/alecthomas/kong"
)

// Globals overridden by linker flags on release build.
var (
	version = "v0.0.0"
)

// Errors returned by the Evy tool.
var (
	errBadWriteFlag = errors.New("cannot use -w without files")
	errNotFormatted = errors.New("not formatted")
	errParse        = errors.New("parse error")
)

var (
	//go:embed build-tools/default-embed
	content    embed.FS
	contentDir = "build-tools/default-embed"
)

const description = `
evy is a tool for managing evy source code.
`

type app struct {
	Version kong.VersionFlag `short:"V" help:"Print version information"`
	Run     runCmd           `cmd:"" help:"Run Evy program."`
	Fmt     fmtCmd           `cmd:"" help:"Format Evy files."`
	Serve   serveCmd         `cmd:"" help:"Start Evy server."`

	Tokenize tokenizeCmd `cmd:"" help:"Tokenize evy program" hidden:""`
	Parse    parseCmd    `cmd:"" help:"Parse evy program" hidden:""`
}

func main() {
	kopts := []kong.Option{
		kong.Description(description),
		kong.Vars{"version": version},
	}
	kctx := kong.Parse(&app{}, kopts...)
	kctx.FatalIfErrorf(kctx.Run())
}

type runCmd struct {
	Source             string `arg:"" help:"Source file. Default: stdin." default:"-"`
	SkipSleep          bool   `help:"Skip evy sleep command." env:"EVY_SKIP_SLEEP"`
	SVGOut             string `help:"Output drawing to SVG file. Stdout: -." placeholder:"FILE"`
	SVGStyle           string `help:"Style of top-level SVG element." placeholder:"STYLE"`
	NoAssertionSummary bool   `short:"s" help:"Do not print assertion summary, only report failed assertion(s)."`
	FailFast           bool   `help:"Stop execution on first failed assertion."`
}

type fmtCmd struct {
	Write bool     `short:"w" help:"Update .evy file." xor:"mode"`
	Check bool     `short:"c" help:"Check if already formatted." xor:"mode"`
	Files []string `arg:"" optional:"" help:"Source files. Default: stdin."`
}

type serveCmd struct {
	Export exportCmd `help:"Export embedded content." cmd:""`
	Start  startCmd  `help:"Start web server, default for \"evy serve\"." cmd:"" default:"withargs"`
}

type exportCmd struct {
	Dir   string `help:"Directory to export embedded content to" short:"d" arg:"" placeholder:"DIR"`
	Force bool   `help:"Use non-empty directory" short:"f"`
}

type startCmd struct {
	Port          int    `help:"Port to listen on" short:"p" default:"8080" env:"EVY_PORT"`
	AllInterfaces bool   `help:"Listen only on all interfaces not just localhost"  short:"a" env:"EVY_ALL_INTERFACES"`
	Dir           string `help:"Directory to serve instead of embedded content" short:"d" type:"existingdir" placeholder:"DIR"`
	Root          string `help:"Directory to use as root for serving, subdirectory of DIR if given, eg \"play\", \"docs\"" placeholder:"DIR"`
}

type tokenizeCmd struct {
	Source string `arg:"" help:"Source file. Default: stdin" default:"-"`
}

type parseCmd struct {
	Source string `arg:"" help:"Source file. Default: stdin" default:"-"`
}

// Run implements the `evy run` CLI command, called by the Kong API.
func (c *runCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}

	rt := cli.NewRuntime(c.runtimeOptions()...)

	eval := evaluator.NewEvaluator(rt)
	eval.AssertInfo.NoAssertionSummary = c.NoAssertionSummary
	eval.AssertInfo.FailFast = c.FailFast
	evyErr := eval.Run(string(b))
	if !errors.As(evyErr, &parser.Errors{}) {
		// even if there was an evaluator error, we want to write as much of the SVG that was produced.
		err = c.writeSVG(rt)
	}
	handleEvyErr(cmp.Or(evyErr, err))
	return nil
}

func (c *runCmd) runtimeOptions() []cli.Option {
	opts := []cli.Option{cli.WithSkipSleep(c.SkipSleep)}
	if c.SVGOut != "" {
		opts = append(opts, cli.WithSVG(c.SVGStyle))
	}
	return opts
}

func (c *runCmd) writeSVG(rt *cli.Runtime) error {
	if c.SVGOut == "" {
		return nil
	}
	var w io.Writer
	if c.SVGOut == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(c.SVGOut)
		if err != nil {
			return fmt.Errorf("cannot create SVG output file: %w", err)
		}
		w = f
		defer f.Close() //nolint:errcheck
	}
	return rt.WriteSVG(w)
}

func handleEvyErr(err error) {
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
	builtins := evaluator.BuiltinDecls()
	prog, err := parser.Parse(in, builtins)
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

func (c *startCmd) Run() error {
	serverRoot, err := c.rootDir()
	if err != nil {
		return err
	}
	http.Handle("/", http.FileServer(serverRoot))
	addr := listenAddr(c.Port, c.AllInterfaces)
	listenAddr, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := &http.Server{ReadHeaderTimeout: 30 * time.Second}
	fmt.Printf("Starting HTTP server on %s\n", listenAddrURL(listenAddr.Addr()))
	return server.Serve(listenAddr)
}

func (c *startCmd) rootDir() (http.FileSystem, error) {
	if c.Dir != "" {
		dir := filepath.Join(c.Dir, c.Root)
		return http.Dir(dir), nil
	}
	dir := filepath.Join(contentDir, c.Root)
	root, err := fs.Sub(content, dir)
	if err != nil {
		return nil, err
	}
	return http.FS(root), nil
}

func (c *exportCmd) Run() error {
	if err := validateExportDir(c.Force, c.Dir); err != nil {
		return err
	}
	fsys, err := fs.Sub(content, contentDir)
	if err != nil {
		return err
	}
	return syncFS(c.Dir, fsys)
}

func validateExportDir(force bool, name string) error {
	f, err := os.Open(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close() //nolint:errcheck // don't care about close failing on read-only files

	_, err = f.ReadDir(1)
	if errors.Is(err, io.EOF) { // empty
		return nil
	}
	if err == nil { // not empty and not force
		if force {
			return nil
		}
		//nolint:goerr113 // dynamic errors in package main is ok
		return fmt.Errorf("%q is not empty, use --force", name)
	}
	return err
}

// syncFS writes the contents of source to a directory on disk. It will
// overwrite any files in the destination directory that have the same name in
// the source filesystem. It will not touch any existing files in the
// destination hierarchy that do not exist in the source filesystem.
func syncFS(destDir string, source fs.FS) error {
	return fs.WalkDir(source, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, filename)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o777)
		}
		sf, err := source.Open(filename)
		if err != nil {
			return err
		}
		defer sf.Close() //nolint:errcheck // don't care about close failing on read-only files
		df, err := os.Create(dest)
		if err != nil {
			return err
		}
		_, err = io.Copy(df, sf)
		if err != nil {
			df.Close() //nolint:errcheck,gosec // we're returning the more important error
			return err
		}
		return df.Close()
	})
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
	builtins := evaluator.BuiltinDecls()
	ast, err := parser.Parse(string(b), builtins)
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

func listenAddr(port int, allInterfaces bool) string {
	if allInterfaces {
		return fmt.Sprintf(":%d", port)
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

func listenAddrURL(address net.Addr) string {
	addr, ok := address.(*net.TCPAddr)
	if !ok {
		return "<unknown address>"
	}
	if addr.IP.IsLoopback() {
		return fmt.Sprintf("http://localhost:%d", addr.Port)
	}
	if addr.IP.IsUnspecified() {
		if h, err := os.Hostname(); err == nil {
			hostPort := net.JoinHostPort(h, strconv.Itoa(addr.Port))
			return "http://" + hostPort
		}
	}
	return "http://" + addr.String()
}
