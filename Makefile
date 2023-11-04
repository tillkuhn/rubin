SHELL=/bin/bash -e -o pipefail
PWD = $(shell pwd)

# constants
GOLANGCI_VERSION = 1.55.1
DOCKER_REPO = rubin
DOCKER_TAG = latest

# defaults for run targets, to overwrite use
# TOPIC="something.else" make run
TOPIC ?= "public.hello"

# customization
.DEFAULT_GOAL = help
export KAFKA_REST_ENDPOINT ?= $(shell test -f pkg/rubin/.test-int-options.yaml && grep rest_endpoint pkg/rubin/.test-int-options.yaml|cut -d: -f2-|xargs || echo "https://localhost:443")
export KAFKA_CLUSTER_ID ?= $(shell test -f pkg/rubin/.test-int-options.yaml && grep cluster_id pkg/rubin/.test-int-options.yaml|cut -d: -f2-|xargs || echo "")
export KAFKA_PRODUCER_API_KEY ?= $(shell test -f pkg/rubin/.test-int-options.yaml && grep api_key pkg/rubin/.test-int-options.yaml|cut -d: -f2-|xargs || echo "")
export KAFKA_PRODUCER_API_SECRET ?= $(shell test -f pkg/rubin/.test-int-options.yaml && grep api_secret pkg/rubin/.test-int-options.yaml|cut -d: -f2-|xargs || echo "")
#export KAFKA_DUMP_MESSAGES ?= $(shell test -f pkg/rubin/.test-int-options.yaml && grep dump_messages pkg/rubin/.test-int-options.yaml|cut -d: -f2-|xargs || echo "")
export KAFKA_LOG_LEVEL ?= $(shell test -f pkg/rubin/.test-int-options.yaml && grep log_level pkg/rubin/.test-int-options.yaml|cut -d: -f2-|xargs || echo "")

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

build: out/bin ## Builds all binaries

GO_BUILD = mkdir -pv "$(@)" && go build -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)' -X 'main.date=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")'" -o "$(@)" ./...
.PHONY: out/bin
out/bin:
	$(GO_BUILD)

GOLANGCI_LINT = bin/golangci-lint-$(GOLANGCI_VERSION)
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b bin v$(GOLANGCI_VERSION)
	@mv bin/golangci-lint "$(@)"

lint: fmt $(GOLANGCI_LINT) download ## Lints all code with golangci-lint
	$(GOLANGCI_LINT) run --fix

lint-reports: out/lint.xml

.PHONY: out/lint.xml
out/lint.xml: $(GOLANGCI_LINT) out download
	$(GOLANGCI_LINT) run ./... --out-format checkstyle | tee "$(@)"

#@go test $(ARGS) ./...
test: ## Runs all tests  (with colorized output support if gotest is installed)
	@if hash gotest 2>/dev/null; then \
	  gotest -v -coverpkg=./... -coverprofile=coverage.out ./...; \
  	else go test -v -coverpkg=./... -coverprofile=coverage.out ./...; fi

test-int: ## Run integration test with tag //go:build integration
	go test --tags=integration ./...

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

clean: ## Cleans up everything
	@rm -rf bin out

docker: ## Builds docker image
	docker buildx build -t $(DOCKER_REPO):$(DOCKER_TAG) .

ci: lint-reports test-reports ## Executes lint and test and generates reports

update: ## Update all go dependencies
	@go get -u all

.PHONY: patch
patch: ## Create Patch Release
	@if hash semtag 2>/dev/null; then \
		semtag final -s patch; \
  	else echo "This target requires semtag, download from https://github.com/nico2sh/semtag"; fi

.PHONY: help
help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort
	@echo ''

run: fmt ## Run the app with JSON String Message
	KAFKA_DUMP_MESSAGES=true go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" \
	./cmd/rubin/main.go -v debug -topic $(TOPIC) -record '{"message":"Hello Franz!"}' \
	-source "rubin/makefile" -subject "my.subject" -header="day=$(shell date +%A)" -header "header2=yeah" -ce

run-help: fmt ## Run the app and display app helm
	@go run -ldflags="-w -s -X 'main.version=$(shell git describe --tags --abbrev=0)' -X 'main.commit=$(shell git rev-parse --short HEAD)'" ./cmd/rubin/main.go -help

