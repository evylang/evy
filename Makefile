# Run `make help` to display help
.DEFAULT_GOAL := $(or $(EVY_DEFAULT_GOAL),help)

# --- Global -------------------------------------------------------------------
O = out
COVERAGE = 80
VERSION ?= $(shell git describe --tags --dirty  --always)
GOFILES = $(shell find . -name '*.go')

all: build test lint tiny test-tiny check-coverage sh-lint check-prettier check-evy-fmt frontend ## Build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

ci: clean check-uptodate all ## Full clean build and up-to-date checks as run on CI

check-uptodate: tidy fmt doc
	test -z "$$(git status --porcelain)" || { git status; false; }

clean:: ## Remove generated files
	-rm -rf $(O)

.PHONY: all check-uptodate ci clean

# --- Build --------------------------------------------------------------------
GO_LDFLAGS = -X main.version=$(VERSION)
CMDS = .

build: | $(O) ## Build evy binaries
	go build -o $(O) -ldflags='$(GO_LDFLAGS)' $(CMDS)

## Build and install binaries in $GOBIN
install:
	go install -ldflags='$(GO_LDFLAGS)' $(CMDS)

# Use `go version` to ensure the right go version is installed when using tinygo.
go-version:
	go version

## Build with tinygo targeting wasm
# optimise for size, see https://www.fermyon.com/blog/optimizing-tinygo-wasm
tiny: go-version | $(O)
	GOOS=wasip1 GOARCH=wasm tinygo build -o $(O)/evy-unopt.wasm -no-debug -ldflags='$(GO_LDFLAGS)' -stack-size=512kb ./pkg/wasm
	wasm-opt -O3 $(O)/evy-unopt.wasm -o frontend/evy.wasm
	cp -f $$(tinygo env TINYGOROOT)/targets/wasm_exec.js frontend/
	echo '{ "version": "$(VERSION)" }' | jq > frontend/version.json

## Tidy go modules with "go mod tidy"
tidy:
	go mod tidy

## Format all go files with gofumpt, a stricter gofmt
fmt:
	gofumpt -w $(GOFILES)

clean::
	-rm -f frontend/evy.wasm
	-rm -f frontend/wasm_exec.js
	-rm -f frontend/version.json

.PHONY: build go-version install tidy tiny

# --- Test ---------------------------------------------------------------------
COVERFILE = $(O)/coverage.txt

## Run non-tinygo tests and generate a coverage file
test: | $(O)
	go test -coverprofile=$(COVERFILE) ./...

## Run tinygo tests
test-tiny: go-version | $(O)
	tinygo test ./...

## Check that test coverage meets the required level
check-coverage: test
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

## Show test coverage in your browser
cover: test
	go tool cover -html=$(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ {print; if ($$3 < $(COVERAGE)) exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: check-coverage cover test test-tiny

# --- Lint ---------------------------------------------------------------------
EVY_FILES = $(shell find frontend/samples -name '*.evy')

## Lint go source code
lint:
	golangci-lint run

## Format evy sample code
evy-fmt:
	go run . fmt --write $(EVY_FILES)

check-evy-fmt:
	go run . fmt --check $(EVY_FILES)

.PHONY: check-evy-fmt evy-fmt lint

# --- Docs ---------------------------------------------------------------------
doc: doctest godoc toc usage

DOCTEST_CMD = ./scripts/doctest.awk $(md) > $(O)/out.md && mv $(O)/out.md $(md)
DOCTESTS = docs/builtins.md docs/spec.md
doctest: install
	$(foreach md,$(DOCTESTS),$(DOCTEST_CMD)$(nl))

TOC_CMD = ./scripts/toc.awk $(md) > $(O)/out.md && mv $(O)/out.md $(md)
TOCFILES = docs/builtins.md docs/spec.md
toc:
	$(foreach md,$(TOCFILES),$(TOC_CMD)$(nl))

USAGE_CMD = ./scripts/gencmd.awk $(md) > $(O)/out.md && mv $(O)/out.md $(md)
USAGEFILES = docs/usage.md
usage: install
	$(foreach md,$(USAGEFILES),$(USAGE_CMD)$(nl))

GODOC_CMD = ./scripts/gengodoc.awk $(filename) > $(O)/out.go && mv $(O)/out.go $(filename)
GODOCFILES = main.go
godoc: install
	$(foreach filename,$(GODOCFILES),$(GODOC_CMD)$(nl))

.PHONY: doc doctest godoc toc usage

# --- frontend -----------------------------------------------------------------
NODEPREFIX = .hermit/node
NODELIB = $(NODEPREFIX)/lib

## Serve frontend on free port
serve:
	servedir frontend

## Format code with prettier
prettier: | $(NODELIB)
	npx --prefix $(NODEPREFIX) -y prettier --write .

## Ensure code is formatted with prettier
check-prettier: | $(NODELIB)
	npx --prefix $(NODEPREFIX) -y prettier --check .

$(NODELIB):
	@mkdir -p $@

.PHONY: check-prettier prettier serve

# --- deploy -----------------------------------------------------------------

## Deploy to live channel on firebase prod, use with care!
## `firebase login` for first time local usage
deploy-prod: tiny
	./scripts/firebase-deploy prod live

## Deploy to live channel on firebase stage.
## `firebase login` for first time local usage
deploy-stage: tiny
	./scripts/firebase-deploy stage live

## Deploy to dev (or other) channel on firebase stage.
## `firebase login` for first time local usage
deploy: tiny
	./scripts/firebase-deploy stage

.PHONY: deploy deploy-prod deploy-stage

# --- scripts ------------------------------------------------------------------
SCRIPTS = scripts/firebase-deploy .github/scripts/app_token

## Lint script files with shellcheck and shfmt
sh-lint:
	shellcheck $(SCRIPTS)
	shfmt --diff $(SCRIPTS)

## Format script files
sh-fmt:
	shfmt --write $(SCRIPTS)

.PHONY: sh-fmt sh-lint

# --- Release -------------------------------------------------------------------
release: nexttag ## Tag and release binaries for different OS on GitHub release
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
	sub(/:$$/, "", $$1)
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
