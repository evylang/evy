//go:build !tinygo

// Evy parses, runs, and formats Evy source code.
//
// Evy on the command line supports all [built-in functions] except for
// graphics functions and event handlers. Only the web interface on
// [play.evy.dev] supports graphics and events.
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
//	  serve export <dir>
//	    Export embedded content.
//
//	  serve start
//	    Start web server, default for "evy serve".
//
//	Run "evy <command> --help" for more information on a command.
//
// [built-in functions]: https://github.com/evylang/evy/blob/main/docs/builtins.md
// [play.evy.dev]: https://play.evy.dev
package main

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"evylang.dev/evy/pkg/evaluator"
	"evylang.dev/evy/pkg/lexer"
	"evylang.dev/evy/pkg/parser"
	"github.com/alecthomas/kong"
)

var (
	version         = "v0.0.0"
	errBadWriteFlag = errors.New("cannot use -w without files")
	errNotFormatted = errors.New("not formatted")
	errParse        = errors.New("parse error")
)

//go:embed out/embed
var content embed.FS

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
	Source    string `arg:"" help:"Source file. Default stdin" default:"-"`
	SkipSleep bool   `help:"skip evy sleep command" env:"EVY_SKIP_SLEEP"`
}

type fmtCmd struct {
	Write bool     `short:"w" help:"update .evy file" xor:"mode"`
	Check bool     `short:"c" help:"check if already formatted" xor:"mode"`
	Files []string `arg:"" optional:"" help:"Source files. Default stdin"`
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
	eval := evaluator.NewEvaluator(rt)
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
	dir := filepath.Join("out/embed", c.Root)
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
	fsys, err := fs.Sub(content, "out/embed")
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
			return os.MkdirAll(dest, 0o777) //nolint:gosec
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
			return fmt.Sprintf("http://%s", hostPort)
		}
	}
	return "http://" + addr.String()
}
