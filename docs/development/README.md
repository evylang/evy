# Getting started with local development

The development process for the [Evy git repository] has been tried and tested
on Linux and MacOS systems. For Windows a [WSL] setup is recommended.

To build the Evy source code, [clone] the repository and
[activate Hermit] in your terminal. Then, build the sources with

    make all

You can list make targets, execute a full CI run or serve the frontend locally
with

    make help
    make ci
    make serve

[Evy git repository]: https://github.com/evylang/evy/
[WSL]: https://learn.microsoft.com/en-us/windows/wsl/about
[clone]: https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository
[activate Hermit]: https://cashapp.github.io/hermit/usage/get-started/?h=activating#activating-an-environment

## Hermit

The tools used in the Evy repository, such as Make, Go and Node, are
automatically downloaded by [Hermit] when needed. Hermit ensures that
developers on Mac, Linux, and GitHub Actions CI use the same version of the
same tools. Cloning this repo is the only installation step necessary.

There are two ways to use the tools in the Evy repository. You can
either prefix them with `bin/`, for example `bin/make all`. Or, you can
activate Hermit in your shell with

    . ./bin/activate-hermit

This will add the tools to your path, so you can use them without having
to prefix them with `bin/`.

You can auto-activate Hermit when changing into the `evy` source
directory by installing [Hermit shell hooks] with

    hermit shell-hooks

[Hermit]: https://cashapp.github.io/hermit
[Hermit shell hooks]: https://cashapp.github.io/hermit/usage/shell/#shell-hooks

## Evy tool chain

The Evy toolchain is written in [Go] and built using the Go and
[TinyGo] compilers. TinyGo targets [WebAssembly], which allows Evy source
code to be parsed and executed in a web browser. However, there is also an
Evy Command Line Interface that can be used to compile and run Evy source
code on your local machine.

A good starting point for understanding the components of the toolchain is
looking at its [usage](../usage.md) documentation. The `evy` toolchain is a set
of tools for parsing, running, and formatting Evy code.

GopherconAU talk as overview, Thorston ball books.

[Go]: https://go.dev
[TinyGo]: https://tinygo.org
[WebAssembly]: https://webassembly.org

<!-- TODO:
### Web Frontend

There is no build process to create the frontend.

- No build process – viewing source should be educational.
- Keep dependencies minimal.
- Provide simple demos for "components", ex. header menu, sidebar, evy editor .

Setting up Deployment Previews [link to Firebase setup guide], for established
members we will provide repo write access.

Relevant make targets:

```sh
make serve
make e2e
make e2e-diff
make snaps
make e2e USE_DOCER=1
make snaps USE_DOCER=1
make docs
make prettier
make style
make deploy
```

Snapshot tests with playwright.

### Build tools

- site-gen
- md
- firebase-deploy
- doctest

### Firebase setup, working with your own fork.
-->
