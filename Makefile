# CI runs `make test`. Local dev uses the mise tasks in mise.toml.

test:
	@echo "+ $@"
	@go test ./... -v
.PHONY: test
