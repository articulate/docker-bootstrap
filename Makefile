PKG_LIST := $(shell go list ./... | grep -v /vendor/)

help:
	@echo "+ $@"
	@grep -Eh '(^[a-zA-Z_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m## /[33m/'
.PHONY: help

##
## Build
## ----------------------------------------------------------------------------

build: ## Build binary
	@echo "+ $@"
	@goreleaser build --snapshot --clean
.PHONY: build

generate: ## Autogenerate code and resources
	@echo "+ $@"
	@go generate ${PKG_LIST}
.PHONY: generate

clean: ## Cleanup build artifacts
	@echo "+ $@"
	@rm -rf dist/
.PHONY: clean

beta: ## Set beta release
	@echo "+ $@"
	@jq '. * {packages: {".": {versioning: "prerelease", prerelease: true}}}' release-please-config.json > release-please-config.json.new
	@mv release-please-config.json.new release-please-config.json
	@git add release-please-config.json
	@git commit -m "chore(release): set to beta"
.PHONY: beta

stable: ## Set stable release
	@echo "+ $@"
	@jq '. * {packages: {".": {versioning: "default", prerelease: false}}}' release-please-config.json > release-please-config.json.new
	@mv release-please-config.json.new release-please-config.json
	@git add release-please-config.json
	@git commit -m "chore(release): promote to stable" -m "Release-As: $$(git describe --tags `git rev-list --tags --max-count=1` | sed -E 's/-.*$$//')"
.PHONY: stable

##
## Development
## ---------------------------------------------------------------------------

init: .git/hooks/pre-commit .git/hooks/commit-msg ## Initialize development environment
	@echo "+ $@"
.PHONY: init

.git/hooks/pre-commit:
ifneq (, $(shell which prek))
	@prek install --hook-type pre-commit
else
	@pre-commit install --install-hooks --hook-type pre-commit
endif

.git/hooks/commit-msg:
ifneq (, $(shell which prek))
	@prek install --hook-type commit-msg
else
	@pre-commit install --install-hooks --hook-type commit-msg
endif

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
	@golangci-lint fmt
	@golangci-lint run --fix
.PHONY: format
