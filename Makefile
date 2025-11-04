# Makefile for Go Projects based on https://github.com/SchwarzIT/go-template + https://github.com/ksoclabs/kbom

## difference := and = https://stackoverflow.com/a/65850656/4292075
## XX = $(shell date) # date is executed every time XX is used, ':=" is executed only once, and ?= only if unset
MODULE_NAME:=$(go list -m)
SHELL=/bin/bash -e -o pipefail
PWD = $(shell pwd)

# constants
GOLANGCI_VERSION = 2.2.1 # run `golangci-lint migrate` to migrate from 1.x
DOCKER_REPO = rubin
DOCKER_TAG = latest

.DEFAULT_GOAL = help

# Customization for this particular project

## default target topic for run targets, to overwrite use
## TOPIC="something.else" make run (...)
TOPIC ?= "public.hello"
## default location of environment file (if not already set, e.g. in .bashrc)
RUBIN_ENV_FILE ?= ".env"  #  "~/.env" for home dir is supported


all: git-hooks  tidy ## Initializes all tools

out:
	@mkdir -p out

git-hooks:
	@git config --local core.hooksPath .githooks/

download: ## Downloads the dependencies
	@go mod download

tidy: ## Cleans up go.mod and go.sum
	@go mod tidy

fmt: ## Formats all code with go fmt
	@go fmt ./...

test-build: ## Tests whether the code compiles
	@go build -o /dev/null ./...

build: out/bin ## Builds all binaries to out/bin

# $@ is the name of the target being generated, and $< the first prerequisite (usually src file).
# in this case, target would be out/bin. See https://stackoverflow.com/a/3220288/4292075
GO_BUILD = mkdir -pv "$(@)" && go build -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)' -X 'main.date=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")'" -o "$(@)" ./...
.PHONY: out/bin
out/bin:
	$(GO_BUILD)
	ls -l out/bin

GOLANGCI_LINT = bin/golangci-lint-$(GOLANGCI_VERSION)
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b bin v$(GOLANGCI_VERSION)
	@mv bin/golangci-lint "$(@)"

lint: fmt $(GOLANGCI_LINT) download ## Lints all code with golangci-lint
	$(GOLANGCI_LINT) run --fix

lint-reports: out/lint.xml

# https://goreleaser.com/cmd/goreleaser_build/ --single-target means ignore config, use only current GOOS and GOARCH
# example binary=dist/rubin_darwin_amd64_v1/rubin
.PHONY: build-goreleaser
build-goreleaser: ## Build binary via goreleaser
	goreleaser build --clean --single-target --skip validate

.PHONY: out/lint.xml
out/lint.xml: $(GOLANGCI_LINT) out download
	$(GOLANGCI_LINT) run ./... --out-format checkstyle | tee "$(@)"

#@go test $(ARGS) ./...
test: ## Runs all tests  (with colorized output support if gotest is installed)
	@if hash gotest 2>/dev/null; then \
	  gotest -v -coverpkg=./... -coverprofile=coverage.out ./...; \
  	else go test -v -coverpkg=./... -coverprofile=coverage.out ./...; fi

.PHONY: test-int
test-int: ## Run integration tests tagged as '//go:build integration'
	go test --tags=integration ./...

test-all: test test-int lint ## Run unit and integration tests followed by lint

coverage: out/report.json ## Displays coverage per func on cli
	go tool cover -func=out/cover.out
	@go tool cover -func cover.out | grep "total:"

coverage-html: out/report.json ## Displays the coverage results in the browser
	@echo Creating HTML coverage report
	go tool cover -html=out/cover.out

test-reports: out/report.json

.PHONY: out/report.json
out/report.json: out
	@go test -count 1 ./... -coverprofile=out/cover.out --json | tee "$(@)"

clean: ## Cleans up everything (bin and out folders)
	@rm -rf bin out

docker: ## Builds docker image
	docker buildx build -t $(DOCKER_REPO):$(DOCKER_TAG) .

ci: lint-reports test-reports ## Executes lint and test and generates reports

update: ## Update all go dependencies
	@go get -u all

.PHONY: help
help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort
	@echo ''

.PHONY: patch
patch: ## Create Patch Release
	@if hash semtag 2>/dev/null; then \
		semtag final -s patch; \
  	else echo "This target requires semtag, download from https://github.com/nico2sh/semtag"; fi

.PHONY: minor
minor: ## Create Minor Release
	@if hash semtag 2>/dev/null; then \
		semtag final -s minor; \
  	else echo "This target requires semtag, download from https://github.com/nico2sh/semtag"; fi

run: fmt ## Run the app with JSON String Message
	KAFKA_DUMP_MESSAGES=false go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" \
	./cmd/rubin/main.go -env-file $(RUBIN_ENV_FILE) -v debug -topic $(TOPIC) -record '{"message":"Hello Franz!"}' \
	-source "urn:rubin:makefile" -subject "my.subject" -header="day=$(shell date +%A)" -header "header2=yeah" -ce

run-polly: fmt ## Run the experimental polly client
	LOG_LEVEL=debug KAFKA_CONSUMER_MAX_RECEIVE=10 go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" \
	./cmd/polly/main.go -env-file $(RUBIN_ENV_FILE) \
	-topic "ci.events" -handler testdata/event-callback.sh

run-help: fmt ## Run the default rubin app and display app helm
	@go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" ./cmd/rubin/main.go -help

run-polly-help: fmt ## Run the polly app and display app helm
	@go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" ./cmd/polly/main.go -help
