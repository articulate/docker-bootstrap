PKG_LIST := $(shell go list ./... | grep -v /vendor/)
RELEASE ?= docker-consul-template-bootstrap

help:
	@echo "+ $@"
	@grep -Eh '(^[a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m## /[33m/'
.PHONY: help

##
## Build
## ----------------------------------------------------------------------------

build: ## Build binary
	@echo "+ $@"
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(RELEASE)
.PHONY: build

generate: ## Autogenerate code and resources
	@echo "+ $@"
	@go generate ${PKG_LIST}
.PHONY: generate

clean: ## Cleanup build artifacts
	@echo "+ $@"
	@rm -rf dist/
.PHONY: clean

##
## Development
## ---------------------------------------------------------------------------

init: ## Initialize development environment
	@echo "+ $@"
	@pre-commit install --install-hooks --hook-type pre-commit --hook-type commit-msg
.PHONY: init

mod: ## Make sure go.mod is up to date
	@echo "+ $@"
	@go mod tidy
.PHONY: mod

test: ## Run tests
	@echo "+ $@"
	@go test ${PKG_LIST} -v
.PHONY: test

coverage: ## Generate test coverage
	@echo "+ $@"
	@go test ${PKG_LIST} -v -coverprofile=coverage.out
	@go tool cover -html=coverage.out
.PHONY: coverage

lint: ## Lint Go code
	@echo "+ $@"
	@golangci-lint run
.PHONY: lint

format: ## Try to fix linting issues
	@echo "+ $@"
	@golangci-lint run --fix
.PHONY: format
