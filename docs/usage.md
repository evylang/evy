# Command Line Usage

The `evy` toolchain is a set of tools for parsing, running, and formatting Evy
code. It also lets you host the Evy web environment locally. You can
[install](install.md) the toolchain and use it from your command line. While the
command-line interface supports most built-in functions, graphics and event
handlers are currently only available within the web interface at
[play.evy.dev].

The Evy toolchain has three subcommands:

- `evy run`: Parse and run Evy source code.
- `evy fmt`: Format Evy source code.
- `evy serve`: Serve Evy website(s) locally.

You can also get help for each subcommand by running it with the
`--help` flag.

[install]: ../README.md#-installation
[play.evy.dev]: https://play.evy.dev

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

      serve export <dir>
        Export embedded content.

      serve start
        Start web server, default for "evy serve".

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

#### evy serve [start] --help

<!-- gen:evy serve start --help -->

    Usage: evy serve start

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

#### evy serve export --help

<!-- gen:evy serve export --help -->

    Usage: evy serve export <dir>

    Export embedded content.

    Arguments:
      <dir>    Directory to export embedded content to

    Flags:
      -h, --help       Show context-sensitive help.
      -V, --version    Print version information

      -f, --force      Use non-empty directory

<!-- genend -->
