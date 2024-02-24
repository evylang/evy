# Installation

You can install the Evy toolchain locally to run it from your command
line. To learn how to use the Evy toolchain, read the
[usage documentation](docs/usage.md) or run `evy --help`.

## Linux and Windows

Download the [latest release] for your platform, unzip it and add `evy`
to your path.

[latest release]: https://github.com/evylang/evy/releases/latest

## macOS

Use [Homebrew] to install `evy`.

    brew install evylang/tap/evy

[Homebrew]: https://brew.sh/

## From Source

[Clone] the [Evy repository] and [activate Hermit] in your terminal. Then,
install from sources with

    make install

and test with

    evy --version
    evy --help

The installed evy toolchain is located in `<REPO>/out/bin/evy`, you may copy
it to a location in your path.

[Evy repository]: https://github.com/evylang/evy/
[Clone]: https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository
[activate Hermit]: https://cashapp.github.io/hermit/usage/get-started/?h=activating#activating-an-environment
