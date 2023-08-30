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

install: ## Build and install binaries in $GOBIN
	go install -ldflags='$(GO_LDFLAGS)' $(CMDS)

# Use `go version` to ensure the right go version is installed when using tinygo.
go-version:
	go version

# Optimise tinygo output for size, see https://www.fermyon.com/blog/optimizing-tinygo-wasm
tiny: go-version | $(O) ## Build for tinygo / wasm
	GOOS=wasip1 GOARCH=wasm tinygo build -o $(O)/evy-unopt.wasm -no-debug -ldflags='$(GO_LDFLAGS)' -stack-size=512kb ./pkg/wasm
	wasm-opt -O3 $(O)/evy-unopt.wasm -o frontend/evy.wasm
	cp -f $$(tinygo env TINYGOROOT)/targets/wasm_exec.js frontend/

tidy: ## Tidy go modules with "go mod tidy"
	go mod tidy

fmt: ## Format all go files with gofumpt, a stricter gofmt
	gofumpt -w $(GOFILES)

clean::
	-rm -f frontend/evy.wasm
	-rm -f frontend/wasm_exec.js

.PHONY: build go-version install tidy tiny

# --- Test ---------------------------------------------------------------------
COVERFILE = $(O)/coverage.txt

test: | $(O) ## Run non-tinygo tests and generate a coverage file
	go test -coverprofile=$(COVERFILE) ./...

test-tiny: go-version | $(O) ## Run tinygo tests
	tinygo test ./...

check-coverage: test ## Check that test coverage meets the required level
	@go tool cover -func=$(COVERFILE) | $(CHECK_COVERAGE) || $(FAIL_COVERAGE)

cover: test ## Show test coverage in your browser
	go tool cover -html=$(COVERFILE)

CHECK_COVERAGE = awk -F '[ \t%]+' '/^total:/ {print; if ($$3 < $(COVERAGE)) exit 1}'
FAIL_COVERAGE = { echo '$(COLOUR_RED)FAIL - Coverage below $(COVERAGE)%$(COLOUR_NORMAL)'; exit 1; }

.PHONY: check-coverage cover test test-tiny

# --- Lint ---------------------------------------------------------------------
EVY_FILES = $(shell find frontend/samples -name '*.evy')
lint: ## Lint go source code
	golangci-lint run

evy-fmt: ## Format evy sample code
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
NODELIB = .hermit/node/lib

frontend: tiny | $(O) ## Build frontend, typically iterate with npm and inside frontend
	rm -rf $(O)/public
	cp -r frontend $(O)/public

frontend-serve: frontend ## Build frontend and serve on free port
	servedir $(O)/public

prettier: | $(NODELIB) ## Format code with prettier
	npx -y prettier --write .

check-prettier: | $(NODELIB)  ## Ensure code is formatted with prettier
	npx -y prettier --check .

$(NODELIB):
	@mkdir -p $@

.PHONY: check-prettier frontend frontend-serve prettier

# --- firebase -----------------------------------------------------------------

firebase-deploy-prod: firebase-public ## Deploy to live channel on firebase, use with care!
	./scripts/firebase-deploy live

firebase-deploy: firebase-public ## Deploy to dev (or other) channel on firebase
	./scripts/firebase-deploy

firebase-emulate: firebase-public ## Run firebase emulator for auth, hosting and datastore
	firebase --config firebase/firebase.json emulators:start

firebase-public: frontend
	rm -rf firebase/public
	cp -r $(O)/public firebase

.PHONY: firebase-deploy firebase-deploy-prod firebase-emulate firebase-public

# --- scripts ------------------------------------------------------------------
SCRIPTS = scripts/firebase-deploy .github/scripts/app_token

sh-lint: ## Lint script files with shellcheck and shfmt
	shellcheck $(SCRIPTS)
	shfmt --diff $(SCRIPTS)

sh-fmt:  ## Format script files
	shfmt --write $(SCRIPTS)

.PHONY: sh-fmt sh-lint

# --- Release -------------------------------------------------------------------
release: nexttag ## Tag and release binaries for different OS on GitHub release
	git tag $(NEXTTAG)
	git push origin $(NEXTTAG)
	[ -z "$(CI)" ] || GITHUB_TOKEN=$$(.github/scripts/app_token) || exit 1; \
	goreleaser release --rm-dist

nexttag:
	$(eval NEXTTAG := $(shell $(NEXTTAG_CMD)))

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
	@awk -F ':.*## ' 'NF == 2 && $$1 ~ /^[A-Za-z0-9%_-]+$$/ { printf "$(COLOUR_WHITE)%-25s$(COLOUR_NORMAL)%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

$(O):
	@mkdir -p $@

.PHONY: help

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
