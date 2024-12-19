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
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

	"evylang.dev/evy/pkg/bytecode"
	"evylang.dev/evy/pkg/cli"
	"evylang.dev/evy/pkg/evaluator"
	"evylang.dev/evy/pkg/lexer"
	"evylang.dev/evy/pkg/parser"
	"github.com/alecthomas/kong"
	"golang.org/x/tools/txtar"
)

// Globals overridden by linker flags on release build.
var (
	version = ""
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
	fullBuild  = false
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
	Compile  compileCmd  `cmd:"" help:"Output compiled bytecode of evy program" hidden:""`
}

func main() {
	kopts := []kong.Option{
		kong.Description(description),
		kong.Vars{"version": getVersion()},
	}
	kctx := kong.Parse(&app{}, kopts...)
	kctx.FatalIfErrorf(kctx.Run())
}

func getVersion() string {
	if version == "" {
		version = getBuildInfoVersion()
	}
	if !fullBuild {
		version += "-slim"
	}
	return version
}

func getBuildInfoVersion() string {
	var mainVersion, revision string
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			mainVersion = info.Main.Version
		}
		settings := map[string]string{}
		for _, s := range info.Settings {
			settings[s.Key] = s.Value
		}
		revision = settings["vcs.revision"]
		if revision != "" && settings["vcs.modified"] == "true" {
			revision += "-dirty"
		}
	}
	return cmp.Or(mainVersion, revision, "unknown")
}

type runCmd struct {
	Source        string `arg:"" help:"Source file. Default: stdin." default:"-"`
	SkipSleep     bool   `help:"Skip evy sleep command." env:"EVY_SKIP_SLEEP"`
	SVGOut        string `help:"Output drawing to SVG file. Stdout: -." placeholder:"FILE"`
	SVGStyle      string `help:"Style of top-level SVG element." placeholder:"STYLE"`
	SVGWidth      string `help:"Width of SVG file." placeholder:"WIDTH"`
	SVGHeight     string `help:"Height of SVG file." placeholder:"HEIGHT"`
	NoTestSummary bool   `short:"s" help:"Do not print test summary, only report failed tests."`
	FailFast      bool   `help:"Stop execution on first failed test."`
	Txtar         string `short:"t" help:"Read source from txtar file and select select given filename" placeholder:"MEMBER"`
	RandSeed      int64  `help:"Seed for random number generation (0 means random seed)."`
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

type compileCmd struct {
	Source string `arg:"" help:"Source file. Default: stdin" default:"-"`
}

// Run implements the `evy run` CLI command, called by the Kong API.
func (c *runCmd) Run() error {
	b, err := c.fileBytes()
	if err != nil {
		return err
	}
	rt := cli.NewPlatform(c.platformOptions()...)
	if c.RandSeed != 0 {
		evaluator.RandSource = rand.New(rand.NewSource(c.RandSeed)) //nolint:gosec // not for security
	}
	eval := evaluator.NewEvaluator(rt)
	eval.TestInfo.NoTestSummary = c.NoTestSummary
	eval.TestInfo.FailFast = c.FailFast
	evyErr := eval.Run(string(b))
	if !errors.As(evyErr, &parser.Errors{}) {
		// even if there was an evaluator error, we want to write as much of the SVG that was produced.
		err = c.writeSVG(rt)
	}
	handleEvyErr(cmp.Or(evyErr, err))
	return nil
}

func (c *runCmd) fileBytes() ([]byte, error) {
	if c.Txtar != "" && filepath.Ext(c.Source) != ".txtar" {
		return nil, errors.New("txtar member specified but source file is not a txtar archive")
	}
	b, err := fileBytes(c.Source)
	if err != nil {
		return nil, err
	}
	if c.Txtar != "" {
		archive := txtar.Parse(b)
		for _, file := range archive.Files {
			if file.Name == c.Txtar {
				return file.Data, nil
			}
		}
		return nil, fmt.Errorf("file %q not found in txtar archive", c.Txtar)
	}
	return b, nil
}

func (c *runCmd) platformOptions() []cli.Option {
	opts := []cli.Option{cli.WithSkipSleep(c.SkipSleep)}
	if c.SVGOut != "" {
		opts = append(opts, cli.WithSVG(c.SVGStyle, c.SVGWidth, c.SVGHeight))
	}
	return opts
}

func (c *runCmd) writeSVG(rt *cli.Platform) error {
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
		return formatStdInOut(c.Check)
	}
	for _, filename := range c.Files {
		var err error
		if filepath.Ext(filename) == ".txtar" {
			err = c.fmtTxtarFile(filename)
		} else {
			err = c.fmtEvyFile(filename)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func formatStdInOut(checkOnly bool) error {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	formatted, err := format(b, checkOnly)
	if err != nil {
		return err
	}
	if !checkOnly {
		fmt.Print(formatted)
	}
	return nil
}

func (c *fmtCmd) fmtEvyFile(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	formatted, err := format(b, c.Check)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	if c.Write {
		return writeAtomically([]byte(formatted), filename)
	}
	return nil
}

func (c *fmtCmd) fmtTxtarFile(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	archive := txtar.Parse(b)
	for i, file := range archive.Files {
		if filepath.Ext(file.Name) != ".evy" {
			continue
		}
		out, err := format(file.Data, c.Check)
		if err != nil {
			return err
		}
		archive.Files[i].Data = []byte(out)
	}
	if c.Write {
		return writeAtomically(txtar.Format(archive), filename)
	}
	return nil
}

func writeAtomically(b []byte, filename string) error {
	tempFile, err := os.CreateTemp(filepath.Dir(filename), "evy")
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	if _, err := tempFile.Write(b); err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	if err := os.Rename(tempFile.Name(), filename); err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	return nil
}

func format(b []byte, checkOnly bool) (string, error) {
	in := string(b)
	builtins := evaluator.BuiltinDecls()
	prog, err := parser.Parse(in, builtins)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errParse, truncateError(err))
	}
	out, err := prog.Format(), nil
	if err != nil {
		return "", err
	}
	if checkOnly && in != out {
		return "", errNotFormatted
	}
	return out, nil
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

// Run implements the hidden `evy compile` CLI command, called by the
// Kong API.
func (c *compileCmd) Run() error {
	b, err := fileBytes(c.Source)
	if err != nil {
		return err
	}
	ast, err := parser.Parse(string(b), parser.Builtins{})
	if err != nil {
		return fmt.Errorf("%w: %w", errParse, truncateError(err))
	}
	comp := bytecode.NewCompiler()
	if err := comp.Compile(ast); err != nil {
		return err
	}
	bc := comp.Bytecode()
	fmt.Println("Num globals:", bc.GlobalCount)
	fmt.Println("Num locals:", bc.LocalCount)
	fmt.Println("Constants:")
	for i, c := range bc.Constants {
		fmt.Printf("%d: %v\n", i, c)
	}
	fmt.Println("\nBytecode:")
	fmt.Print(bc.Instructions)

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
