# Command Line Usage

The `evy` toolchain is a set of tools that can be used to parse, run, and
format Evy source code. It can also be used to serve the Evy web contents
locally. You can [install] the Evy toolchain locally and run it from your
command line. The command-line interface for Evy supports all built-in and
graphics functions except for event handlers. Events are currently only
supported within the web interface, such as on [play.evy.dev].

The Evy toolchain has three subcommands:

- `evy run`: Parse and run Evy source code.
- `evy fmt`: Format Evy source code.
- `evy serve`: Serve Evy website(s) locally.

You can also get help for each subcommand by running it with the
`--help` flag.

[install]: ../README.md#-installation
[play.evy.dev]: https://play.evy.dev

### evy --help

<!-- gen:evy --help -->

    Usage: evy <command> [flags]

    evy is a tool for managing evy source code.

    Flags:
      -h, --help       Show context-sensitive help.
      -V, --version    Print version information

    Commands:
      run [<source>] [flags]
        Run Evy program.

      fmt [<files> ...] [flags]
        Format Evy files.

      serve export <dir> [flags]
        Export embedded content.

      serve start [flags]
        Start web server, default for "evy serve".

    Run "evy <command> --help" for more information on a command.

<!-- genend -->

### evy run --help

<!-- gen:evy run --help -->

    Usage: evy run [<source>] [flags]

    Run Evy program.

    Arguments:
      [<source>]    Source file. Default: stdin.

    Flags:
      -h, --help                    Show context-sensitive help.
      -V, --version                 Print version information

          --skip-sleep              Skip evy sleep command ($EVY_SKIP_SLEEP).
          --svg-out=FILE            Output drawing to SVG file. Stdout: -.
          --svg-style=STYLE         Style of top-level SVG element.
      -s, --no-assertion-summary    Do not print assertion summary, only report
                                    failed assertion(s).
          --fail-fast               Stop execution on first failed assertion.

<!-- genend -->

### evy fmt --help

<!-- gen:evy fmt --help -->

    Usage: evy fmt [<files> ...] [flags]

    Format Evy files.

    Arguments:
      [<files> ...]    Source files. Default: stdin.

    Flags:
      -h, --help       Show context-sensitive help.
      -V, --version    Print version information

      -w, --write      Update .evy file.
      -c, --check      Check if already formatted.

<!-- genend -->

### evy serve [start] --help

<!-- gen:evy serve start --help -->

    Usage: evy serve start [flags]

    Start web server, default for "evy serve".

    Flags:
      -h, --help              Show context-sensitive help.
      -V, --version           Print version information

      -p, --port=8080         Port to listen on ($EVY_PORT)
      -a, --all-interfaces    Listen only on all interfaces not just localhost
                              ($EVY_ALL_INTERFACES)
      -d, --dir=DIR           Directory to serve instead of embedded content
          --root=DIR          Directory to use as root for serving, subdirectory of
                              DIR if given, eg "play", "docs"

<!-- genend -->

### evy serve export --help

<!-- gen:evy serve export --help -->

    Usage: evy serve export <dir> [flags]

    Export embedded content.

    Arguments:
      <dir>    Directory to export embedded content to

    Flags:
      -h, --help       Show context-sensitive help.
      -V, --version    Print version information

      -f, --force      Use non-empty directory

<!-- genend -->
