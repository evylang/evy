# Evy

[![Discord Chat](https://img.shields.io/badge/discord-chat-414eed?style=flat-square&logo=discord&logoColor=white)](https://discord.evy.dev)
[![GitHub Build](https://img.shields.io/github/actions/workflow/status/evylang/evy/prod.yaml?style=flat-square&branch=main&logo=github)](https://github.com/evylang/evy/actions/workflows/prod.yaml?query=branch%3Amain)
[![Go Reference](https://pkg.go.dev/badge/evylang.dev/evy.svg)](https://pkg.go.dev/evylang.dev/evy)
[![GitHub Sponsorship](https://img.shields.io/badge/sponsor-%E2%99%A5-eb5c95?style=flat-square&logo=github&logoColor=white)](https://github.com/sponsors/evylang)

Evy is a simple programming language, made to learn coding. [Try it out].

Evy bridges the gap between block-based languages like [Scratch] and
conventional languages like Python or JavaScript. It has a minimalist
syntax with fewer special characters and advanced concepts than most
programming languages. Evy has a small set of
[built-in functions](docs/builtins.md) that are easy to understand and
remember, but it still is powerful enough for user interaction, games,
and animations.

[Try it out]: https://play.evy.dev
[Scratch]: https://scratch.mit.edu/

## 🌱 Getting Started

You can try Evy online at [play.evy.dev] and browse the examples.
Alternatively you can [install](#-installation) the `evy` toolchain
locally.

Here is a "hello world" program in Evy

    print "Hello World!"

<details>
  <summary>Screencast: How to code a simple animation in Evy.</summary>

![Coding evy](docs/img/purple-dot.gif)

[Animation source code]

</details>

[play.evy.dev]: https://play.evy.dev
[Animation source code]: https://play.evy.dev#content=H4sIAAAAAAAAEzWLwQqAIBBE7/sVg/fSiC6BHyO2B0FXWazvz4qGGXjMMKVejM09thaRpbNSrLkqTDu1ZTak2D1WRzFpzAwlqoIgqYTOhKHRBn1J4UcmuHn5lv/CctAN/HT8mWwAAAA=

## 📖 Documentation

Here are some resources for learning more about Evy:

- [Syntax by Example](docs/syntax-by-example.md): A collection of examples that illustrate the Evy syntax.
- [Built-in Documentation](docs/builtins.md): Details on built-in functions and events in Evy.
- [Language Specification](docs/spec.md): A formal definition of the Evy syntax and language.
- [Interactive Labs]: A guided tour with challenges that teaches you programming using Evy.

For questions and discussions, join the [Evy community] on Discord.

[Evy community]: https://discord.evy.dev
[Interactive Labs]: https://lab.evy.dev

## 📦 Installation

You can install the Evy toolchain locally to run it from your command
line. To learn how to use the Evy toolchain, read the
[usage documentation](docs/usage.md) or run `evy --help`.

### Linux and Windows

Download the [latest release] for your platform, unzip it and add `evy`
to your path.

### macOS

Use [Homebrew] to install `evy`.

    brew install evylang/tap/evy

[latest release]: https://github.com/evylang/evy/releases/latest
[Homebrew]: https://brew.sh/

## 💻 Development

The Evy interpreter is written in [Go] and built using the Go and
[TinyGo] compilers. TinyGo targets [WebAssembly], which allows Evy
source code to be parsed and run in a web browser. The browser
runtime is written in plain JavaScript without the use of frameworks.

To build the Evy source code, [clone] this repository and
[activate Hermit] in your terminal. Then, build the sources with

    make all

You can list make targets, execute a full CI run locally or serve the
frontend locally with

    make help
    make ci
    make serve

<details>
  <summary>Hermit automatically installs tools.</summary>

### Hermit

The tools used in this repository, such as Make, Go and Node, are
automatically downloaded by [Hermit] when needed. Hermit ensures that
developers on Mac, Linux, and GitHub Actions CI use the same version of
the same tools. Cloning this repo is the only installation step
necessary.

There are two ways to use the tools in the Evy repository. You can
either prefix them with `bin/`, for example `bin/make all`. Or, you can
activate Hermit in your shell with

    . ./bin/activate-hermit

This will add the tools to your path, so you can use them without having
to prefix them with `bin/`.

You can auto-activate Hermit when changing into the `evy` source
directory by installing [Hermit shell hooks] with

    hermit shell-hooks

</details>

[Go]: https://go.dev
[TinyGo]: https://tinygo.org
[WebAssembly]: https://webassembly.org
[Clone]: https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository
[activate Hermit]: https://cashapp.github.io/hermit/usage/get-started/?h=activating#activating-an-environment
[Hermit]: https://cashapp.github.io/hermit
[Hermit shell hooks]: https://cashapp.github.io/hermit/usage/shell/#shell-hooks

## 🙏 Thanks

Evy would not be here today without the help of many people.

- [camh]: Thank you for your support, guidance, generosity and endless patience.
- [Jason]: Thank you for donating the Evy website design. It is beautiful!
- [@ckaser]: Thank you for creating the fantastic [easylang] language, which was a major inspiration for Evy.
- [@fcostin]: Thank you for the [Splash of Trig] sample, your wisdom and your willingness to help.
- [@starkcoffee], [@alecthomas], [@loislambeth]: Thank you for your insights, support, and encouragement. I am grateful for your friendship!
- My daughter Mali: Thank you for being keen to learn programming and for testing Evy with me.

[camh]: https://github.com/camh-
[Jason]: https://twitter.com/jasonstrachan
[@ckaser]: https://github.com/ckaser
[easylang]: https://easylang.online/
[@fcostin]: https://github.com/fcostin
[Splash of Trig]: https://play.evy.dev#splashtrig
[@starkcoffee]: https://github.com/starkcoffee
[@loislambeth]: https://github.com/loislambeth
[@alecthomas]: https://github.com/alecthomas
