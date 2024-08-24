# Upgrade Tools and Dependencies

We attempt to upgrade tools and dependencies once a month. This upgrade process
has five parts:

1. [Upgrade Hermitised Tools](#upgrade-hermitised-tools)
2. [Upgrade Go Dependencies](#upgrade-go-dependencies)
3. [Upgrade Frontend Tools](#upgrade-frontend-tools)
4. [Upgrade Playwright](#upgrade-playwright)
5. [Test External URLs](#test-external-urls)

Each step should follow in the listed order and be committed separately. Ensure
`make ci` still passes before continuing with the next step.

[Go Dependencies]: https://go.dev/doc/modules/managing-dependencies
[NPM]: https://www.npmjs.com/

## Upgrade Hermitised Tools

The tools used in this repository, such as Make, Go and Node, are
automatically downloaded by [Hermit] when needed. Hermit ensures that
developers on Mac, Linux, and GitHub Actions CI use the same version of
the same tools.

[Hermit]: https://cashapp.github.io/hermit/

To upgrade all tools managed by Hermit run

    hermit upgrade

Then run

    hermit search --exact TOOL

for tools you suspect might have major version bumps - node, firebase,
openjre and goreleaser had them recently. Major version upgrades or
uninterpretable version upgrades don't get installed automatically. If
needed install specific versions manually with

    hermit install TOOL-x.y.z

Commit changes and ensure the build still works with

    make ci

## Upgrade Go Dependencies

Upgrade all Go package dependencies with

    go get -u ./...
    go mod tidy

and the same again for the `evy/learn` sub-module

    go -C learn get -u ./...
    go -C learn mod tidy

Verify that the Go version specified in the `go.mod` files matches the Go
version that Hermit installs, which has potentially been upgraded in the prior
step. If necessary, change the Go version in the `go.mod` files.

Commit changes and ensure the build still works with

    make ci

## Upgrade Frontend Tools

Upgrade other NPM frontend formatting and linting tools such as [prettier] or
[stylelint]. From the repository root directory run

    npx --prefix .hermit/node -y npm-check-updates --packageFile .hermit/node/package.json -u
    npm --prefix .hermit/node install

Commit changes and ensure the build still works with

    make ci

[prettier]: https://prettier.io/
[stylelint]: https://stylelint.io/

## Upgrade Playwright

### Install new Playwright version

We use [Playwright] for automated end-to-end and browser testing.

From the repository root directory run

    npx --prefix e2e -y npm-check-updates --packageFile e2e/package.json -u
    npm --prefix e2e install
    npx --prefix e2e playwright install

If a new version of Playwright has been installed, also update the Docker image
version of Playwright used in Makefile and on CI.

Find the correct Docker tag in the [Playwright Docker documentation]. Replace it
in the Makefile, for example update:

    PLAYWRIGHT_OCI_IMAGE = mcr.microsoft.com/playwright:v1.46.0-jammy

### Ensure end-to-end tests still pass

Start a local server in one terminal with

    make serve

Run end-to-end tests in another terminal with

    make e2e

If snapshot tests fail and you are certain that the snapshot-diffs are
justified, update snapshots with

    make snaps

The step above is only used for local development. To run Playwright with docker
as GitHub Actions CI does use:

    make e2e USE_DOCKER=1

if snapshots need updating run

    make snaps USE_DOCKER=1

If there are connection errors between Docker and the local development server,
try starting the server with

    make serve SERVEDIR_ALL_INTERFACES=1

Commit changes and ensure the build still works with

    make ci

[Playwright]: https://playwright.dev/
[Playwright Docker documentation]: https://playwright.dev/docs/docker

## Test External URLs

Run

    make test-urls

to check that all external URLs are reachable.
