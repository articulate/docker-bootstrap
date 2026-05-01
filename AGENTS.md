# AGENTS.md

## Critical: Blast Radius

This binary is used as the entrypoint for several Docker images. A bug here
affects every service that depends on it, and rolling out a fix requires
rebuilding and redeploying those images. Be extremely cautious with
changes.

## Env Var Precedence

Vault values overwrite Consul values at the same key. Pre-existing host env
vars are never overwritten. Vault skips the `global/env_vars` and
`services/{name}/env_vars` paths in `dev` and `test` environments (the
`devEnv` guard in `config.go:VaultPaths()`).

## Variable Expansion in Values

Loaded values are passed through `os.Expand` (`env.go:47`). A Consul/Vault
value containing `$FOO` or `${FOO}` will be expanded against other loaded
vars and the host environment. This is intentional but easy to miss — values
with literal `$` must avoid this syntax or they'll be silently rewritten.

## Gotchas

- `scripts/docker-secrets` must be **sourced** (`. $0`), not executed, when
  loading secrets into the current shell. Executing it runs `exec "$@"` and
  replaces the shell.
- Integration tests require Vault KV **v1**. The `tests/setup` script
  explicitly dismounts KV v2 and re-enables v1. If the CI Vault image
  defaults change, tests will silently return empty values.
- `make beta` and `make stable` commit directly to the current branch.
