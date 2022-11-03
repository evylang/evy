# Run `make help` to display help
.DEFAULT_GOAL := help

# --- Global -------------------------------------------------------------------
O = out
COVERAGE = 80
VERSION ?= $(shell git describe --tags --dirty  --always)

all: build test lint tiny test-tiny check-coverage sh-lint frontend ## Build, test, check coverage and lint
	@if [ -e .git/rebase-merge ]; then git --no-pager log -1 --pretty='%h %s'; fi
	@echo '$(COLOUR_GREEN)Success$(COLOUR_NORMAL)'

ci: clean check-uptodate all ## Full clean build and up-to-date checks as run on CI

check-uptodate: tidy
	test -z "$$(git status --porcelain -- go.mod go.sum)" || { git status; false; }

clean:: ## Remove generated files
	-rm -rf $(O)

.PHONY: all check-uptodate ci clean

# --- Build --------------------------------------------------------------------
GO_LDFLAGS = -X main.version=$(VERSION)
CMDS = .

build: | $(O) ## Build reflect binaries
	go build -o $(O) -ldflags='$(GO_LDFLAGS)' $(CMDS)

install: ## Build and install binaries in $GOBIN
	go install -ldflags='$(GO_LDFLAGS)' $(CMDS)

# Use `go version` to ensure the right go version is installed when using tinygo.
go-version:
	go version

# Optimise tinygo output for size, see https://www.fermyon.com/blog/optimizing-tinygo-wasm
tiny: go-version | $(O) ## Build for tinygo / wasm
	tinygo build -o frontend/evy.wasm -target wasm -no-debug -ldflags='$(GO_LDFLAGS)' ./pkg/wasm
	cp -f $$(tinygo env TINYGOROOT)/targets/wasm_exec.js frontend/

tidy: ## Tidy go modules with "go mod tidy"
	go mod tidy

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
lint: ## Lint go source code
	golangci-lint run

.PHONY: lint

# --- frontend -----------------------------------------------------------------
frontend: tiny | $(O) ## Build frontend, typically iterate with npm and inside frontend
	rm -rf $(O)/public
	cp -r frontend $(O)/public

frontend-serve: frontend ## Build frontend and serve on free port
	servedir $(O)/public

.PHONY: frontend frontend-serve

# --- firebase -----------------------------------------------------------------

firebase-deploy-prod: firebase-public ## Deploy to live channel on firebase, use with care!
	./firebase/deploy live

firebase-deploy: firebase-public ## Deploy to dev (or other) channel on firebase
	./firebase/deploy

firebase-emulate: firebase-public ## Run firebase emulator for auth, hosting and datastore
	firebase --config firebase/firebase.json emulators:start

firebase-public: frontend
	rm -rf firebase/public
	cp -r $(O)/public firebase

.PHONY: firebase-deploy firebase-deploy-prod firebase-emulate firebase-public

# --- scripts ------------------------------------------------------------------
SCRIPTS = firebase/deploy .github/scripts/app_token

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
{ git tag --list --merged HEAD --sort=-v:refname; echo v0.0.0; }
| grep -E "^v?[0-9]+.[0-9]+.[0-9]+$$"
| head -n1
| awk -F . '{ print $$1 "." $$2 "." $$3 + 1 }'
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
