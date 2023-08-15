# `evy` Usage

The `evy` toolchain is a set of tools that can be used to compile, run,
and format Evy source code. You can install the Evy toolchain locally
and run it from your command line. The command-line interface for Evy
supports all built-in functions except for graphics functions and event
handlers. Only the web interface on [evy.dev] supports graphics and
events.

The Evy toolchain has two subcommands:

- `evy run`: Compile and run Evy source code.
- `evy fmt`: Format Evy source code.

You can also get help for each subcommand by running it with the
`--help` flag.

[evy.dev]: https://evy.dev

#### evy --help

<!-- gen:evy --help -->

    Usage: evy <command>

    evy is a tool for managing evy source code.

    Flags:
      -h, --help       Show context-sensitive help.
      -V, --version    Print version information

    Commands:
      run [<source>]
        Run Evy program.

      fmt [<files> ...]
        Format Evy files.

    Run "evy <command> --help" for more information on a command.

<!-- genend -->

#### evy run --help

<!-- gen:evy run --help -->

    Usage: evy run [<source>]

    Run Evy program.

    Arguments:
      [<source>]    Source file. Default stdin

    Flags:
      -h, --help          Show context-sensitive help.
      -V, --version       Print version information

          --skip-sleep    skip evy sleep command ($EVY_SKIP_SLEEP)

<!-- genend -->

#### evy fmt --help

<!-- gen:evy fmt --help -->

    Usage: evy fmt [<files> ...]

    Format Evy files.

    Arguments:
      [<files> ...]    Source files. Default stdin

    Flags:
      -h, --help       Show context-sensitive help.
      -V, --version    Print version information

      -w, --write      update .evy file
      -c, --check      check if already formatted

<!-- genend -->
