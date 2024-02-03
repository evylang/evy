# Run `make help` to display help
.DEFAULT_GOAL := $(or $(EVY_DEFAULT_GOAL),help)

# --- Global -------------------------------------------------------------------
O = out
COVERAGE = 80
VERSION ?= $(shell git describe --tags --dirty  --always)
GOFILES = $(shell find . -name '*.go')

## Build, test, check coverage and lint
all: build-go test lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

test: test-go test-tiny test-cli check-coverage

lint: lint-go lint-sh check-prettier check-style check-fmt-evy

## Full clean build and up-to-date checks as run on CI
ci: clean check-uptodate all

check-uptodate: tidy fmt doc
	test -z "$$(git status --porcelain)" || { git status; false; }

## Remove generated files
clean::
	-rm -rf $(O)

.PHONY: all check-uptodate ci test lint clean

# --- Build --------------------------------------------------------------------
GO_LDFLAGS = -X main.version=$(VERSION)
CMDS = .

## Build evy binaries
build-go: embed | $(O)
	go build -o $(O) -ldflags='$(GO_LDFLAGS)' $(CMDS)

## Build and install binaries in $GOBIN
install: embed
	go install -ldflags='$(GO_LDFLAGS)' $(CMDS)

## Build and install slim binaries without embedded frontend in $GOBIN
install-slim: embed-slim
	go install -ldflags='$(GO_LDFLAGS)' $(CMDS)

# Use `go version` to ensure the right go version is installed when using tinygo.
go-version:
	go version

## Build with tinygo targeting wasm
# optimise for size, see https://www.fermyon.com/blog/optimizing-tinygo-wasm
build-tiny: go-version | $(O)
	GOOS=wasip1 GOARCH=wasm tinygo build -o $(O)/evy-unopt.wasm -no-debug -ldflags='$(GO_LDFLAGS)' -stack-size=512kb ./pkg/wasm
	wasm-opt -O3 $(O)/evy-unopt.wasm -o frontend/module/evy.wasm
	cp -f $$(tinygo env TINYGOROOT)/targets/wasm_exec.js frontend/module/
	echo '{ "version": "$(VERSION)" }' | jq > frontend/version.json

## Prepare frontend assets to be embedded into the binary
embed: build-tiny | $(O)
	rm -rf $(O)/embed
	go run ./build-tools/site-gen frontend $(O)/embed

## Prepare slim frontend assets, with placeholder index.html only, to be embedded into the binary
embed-slim: | $(O)
	rm -rf $(O)/embed
	mkdir $(O)/embed
	cp build-tools/embed-slim-index.html $(O)/embed/index.html

## Tidy go modules with "go mod tidy"
tidy:
	go mod tidy

## Format all go files with gofumpt, a stricter gofmt
fmt:
	gofumpt -w $(GOFILES)

clean::
	-rm -f frontend/module/evy.wasm
	-rm -f frontend/module/wasm_exec.js
	-rm -f frontend/version.json

.PHONY: build-go build-tiny embed embed-slim go-version install install-slim tidy

# --- Test ---------------------------------------------------------------------
COVERFILE = $(O)/coverage.txt
EXPORTDIR = $(O)/export-test

## Run non-tinygo tests and generate a coverage file
test-go: embed-slim | $(O)
	go test -coverprofile=$(COVERFILE) ./...

## Test evy CLI
test-cli: build-go
	rm -rf $(EXPORTDIR)
	$(O)/evy serve export $(EXPORTDIR)
	test -f $(EXPORTDIR)/index.html
	test -f $(EXPORTDIR)/play/module/evy.wasm
	test ! -L $(EXPORTDIR)/play/module/evy.wasm

## Run tinygo tests
test-tiny: go-version | $(O)
	tinygo test ./...

## Check that test coverage meets the required level
check-coverage: test-go
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

## Show test coverage in your browser
cover: test-go
	go tool cover -html=$(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ {print; if ($$3 < $(COVERAGE)) exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: check-coverage cover test-cli test-go test-tiny

# --- Lint ---------------------------------------------------------------------
EVY_FILES = $(shell find frontend/play/samples -name '*.evy')

## Lint go source code
lint-go: embed-slim
	golangci-lint run

## Format evy sample code
fmt-evy:
	go run . fmt --write $(EVY_FILES)

check-fmt-evy:
	go run . fmt --check $(EVY_FILES)

.PHONY: check-fmt-evy fmt-evy lint-go

# --- Docs ---------------------------------------------------------------------
doc: doctest godoc toc usage

DOCTEST_CMD = ./build-tools/doctest.awk $(md) > $(O)/out.md && mv $(O)/out.md $(md)
DOCTESTS = docs/builtins.md docs/spec.md
doctest: install-slim
	$(foreach md,$(DOCTESTS),$(DOCTEST_CMD)$(nl))

TOC_CMD = ./build-tools/toc.awk $(md) > $(O)/out.md && mv $(O)/out.md $(md)
TOCFILES = docs/builtins.md docs/spec.md
toc:
	$(foreach md,$(TOCFILES),$(TOC_CMD)$(nl))

USAGE_CMD = ./build-tools/gencmd.awk $(md) > $(O)/out.md && mv $(O)/out.md $(md)
USAGEFILES = docs/usage.md
usage: install-slim
	$(foreach md,$(USAGEFILES),$(USAGE_CMD)$(nl))

GODOC_CMD = ./build-tools/gengodoc.awk $(filename) > $(O)/out.go && mv $(O)/out.go $(filename)
GODOCFILES = main.go
godoc: install-slim
	$(foreach filename,$(GODOCFILES),$(GODOC_CMD)$(nl))

.PHONY: doc doctest godoc toc usage

# --- frontend -----------------------------------------------------------------
NODEPREFIX = .hermit/node
NODELIB = $(NODEPREFIX)/lib

define PLAYWRIGHT_CMD
	npm --prefix e2e ci
	npx --prefix e2e playwright test --config e2e $(PLAYWRIGHT_ARGS)
endef

PLAYWRIGHT_IMG = mcr.microsoft.com/playwright:v1.41.1-jammy
PLAYWRIGHT_CMD_DOCKER = docker run --rm \
  --volume $$(pwd):/work/ -w /work/ \
  --network host --add-host=host.docker.internal:host-gateway \
  --env BASEURL=$(BASEURL) \
  $(PLAYWRIGHT_IMG) /bin/bash -e -c "$(subst $(nl),;,$(PLAYWRIGHT_CMD))"

# BASEURL needs to be in the environment so that `e2e/playwright.config.js`
# can see it when the `e2e` target is called.
# The firebase-deploy script sets BASEURL to the deployment URL on GitHub CI.
SERVEDIR_HOST = localhost
export SERVEDIR_PORT ?= 8080
export BASEURL ?= http://$(SERVEDIR_HOST):$(SERVEDIR_PORT)

## Serve frontend on port 8080 by default to work with e2e target
serve:
	servedir frontend

## Format code with prettier
prettier: | $(NODELIB)
	npx --prefix $(NODEPREFIX) -y prettier --write .

## Ensure code is formatted with prettier
check-prettier: | $(NODELIB)
	npx --prefix $(NODEPREFIX) -y prettier --check .

## Fix CSS files with stylelint
style: | $(NODELIB)
	npm --prefix $(NODEPREFIX) ci
	npx --prefix $(NODEPREFIX) stylelint -c $(NODEPREFIX)/.stylelintrc.json --fix frontend/**/*.css

## Lint CSS files with stylelint
check-style: | $(NODELIB)
	npm --prefix $(NODEPREFIX) ci
	npx --prefix $(NODEPREFIX) stylelint -c $(NODEPREFIX)/.stylelintrc.json frontend/**/*.css

## Install playwright on host system for `e2e` to use.
install-playwright:
	npx --prefix e2e playwright install --with-deps chromium

## Run end-to-end test on host system, could be MacOS, Linux or other
e2e:
	@echo "testing $(BASEURL)"
	$(PLAYWRIGHT_CMD)

## Run end-to-end tests with Docker, used on Linux CI
e2e-docker: SERVEDIR_HOST = host.docker.internal
e2e-docker:
	$(PLAYWRIGHT_CMD_DOCKER)

## Make end-to-end testing golden screenshots for local OS
snaps: PLAYWRIGHT_ARGS = --update-snapshots
snaps:
	$(PLAYWRIGHT_CMD)

## Make end-to-end testing golden screenshots with Docker, used on Linux CI
snaps-docker: SERVEDIR_HOST = host.docker.internal
snaps-docker: PLAYWRIGHT_ARGS = --update-snapshots
snaps-docker:
	$(PLAYWRIGHT_CMD_DOCKER)

$(NODELIB):
	@mkdir -p $@

.PHONY: check-prettier e2e prettier serve

# --- deploy -----------------------------------------------------------------

## Deploy to live channel on firebase prod, use with care!
## `firebase login` for first time local usage
deploy-prod: build-tiny
	./build-tools/firebase-deploy prod live

## Deploy to live channel on firebase stage.
## `firebase login` for first time local usage
deploy-stage: build-tiny
	./build-tools/firebase-deploy stage live

## Deploy to dev (or other) channel on firebase stage.
## `firebase login` for first time local usage
deploy: build-tiny
	./build-tools/firebase-deploy stage

.PHONY: deploy deploy-prod deploy-stage

# --- scripts ------------------------------------------------------------------
SCRIPTS = build-tools/firebase-deploy .github/scripts/app_token

## Lint script files with shellcheck and shfmt
lint-sh:
	shellcheck $(SCRIPTS)
	shfmt --diff $(SCRIPTS)

## Format script files
fmt-sh:
	shfmt --write $(SCRIPTS)

.PHONY: fmt-sh lint-sh

# --- Release -------------------------------------------------------------------
## Tag and release binaries for different OS on GitHub release
# We need to run embed first to generate the full website including evy.wasm
# for embedding in the go binary. goreleaser build hooks cannot be used as
# they run in parallel for each os/arch and cause a race condition.
release: nexttag embed
	git tag $(NEXTTAG)
	git push origin $(NEXTTAG)
	[ -z "$(CI)" ] || GITHUB_TOKEN=$$(.github/scripts/app_token) || exit 1; \
	goreleaser release --clean $(if $(RELNOTES),--release-header=$(RELNOTES))

nexttag:
	$(eval NEXTTAG := $(shell $(NEXTTAG_CMD)))
	$(eval RELNOTES := $(wildcard docs/release-notes/$(NEXTTAG).md))

.PHONY: nexttag release

define NEXTTAG_CMD
{
  { git tag --list --merged HEAD --sort=-v:refname; echo v0.0.0; }
  | grep -E "^v?[0-9]+\.[0-9]+\.[0-9]+$$"
  | head -n 1
  | awk -F . '{ print $$1 "." $$2 "." $$3 + 1 }';
  git diff --name-only @^ | sed -E -n 's|^docs/release-notes/(v[0-9]+\.[0-9]+\.[0-9]+)\.md$$|\1|p';
} | sort --reverse --version-sort | head -n 1
endef

# --- Utilities ----------------------------------------------------------------
COLOUR_NORMAL = $(shell tput sgr0 2>/dev/null)
COLOUR_RED    = $(shell tput setaf 1 2>/dev/null)
COLOUR_GREEN  = $(shell tput setaf 2 2>/dev/null)
COLOUR_WHITE  = $(shell tput setaf 7 2>/dev/null)

help:
	$(eval export HELP_AWK)
	@awk "$${HELP_AWK}" $(MAKEFILE_LIST) | sort | column -s "$$(printf \\t)" -t

$(O):
	@mkdir -p $@

.PHONY: help

# Awk script to extract and print target descriptions for `make help`.
define HELP_AWK
/^## / { desc = desc substr($$0, 3) }
/^[A-Za-z0-9%_-]+:/ && desc {
	sub(/::?$$/, "", $$1)
	printf "$(COLOUR_WHITE)%s$(COLOUR_NORMAL)\t%s\n", $$1, desc
	desc = ""
}
endef

define nl


endef
ifndef ACTIVE_HERMIT
$(eval $(subst \n,$(nl),$(shell bin/hermit env -r | sed 's/^\(.*\)$$/export \1\\n/')))
endif

# Ensure make version is gnu make 3.82 or higher
ifeq ($(filter undefine,$(value .FEATURES)),)
$(error Unsupported Make version. \
	$(nl)Use GNU Make 3.82 or higher (current: $(MAKE_VERSION)). \
	$(nl)Activate 🐚 hermit with `. bin/activate-hermit` and run again \
	$(nl)or use `bin/make`)
endif
