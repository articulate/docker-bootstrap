# CI runs `make test` and releases use `make beta` / `make stable`. Local dev
# uses the mise tasks in mise.toml.

test:
	@echo "+ $@"
	@go test ./... -v
.PHONY: test

beta:
	@echo "+ $@"
	@jq '. * {packages: {".": {versioning: "prerelease", prerelease: true}}}' release-please-config.json > release-please-config.json.new
	@mv release-please-config.json.new release-please-config.json
	@git add release-please-config.json
	@git commit -m "chore(release): set to beta"
.PHONY: beta

stable:
	@echo "+ $@"
	@jq '. * {packages: {".": {versioning: "default", prerelease: false}}}' release-please-config.json > release-please-config.json.new
	@mv release-please-config.json.new release-please-config.json
	@git add release-please-config.json
	@git commit -m "chore(release): promote to stable" -m "Release-As: $$(git describe --tags `git rev-list --tags --max-count=1` | sed -E 's/-.*$$//')"
.PHONY: stable
