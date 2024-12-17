# Docker Consul Template Bootstrap

Load values from Consul and Vault as environment variables.

## Installing

Download the [latest release](https://github.com/articulate/docker-bootstrap/releases),
add it to your image, and set it as your `ENTRYPOINT`.

If you are using Buildkit you can use the `TARGETARCH` arg to `ADD` the correct
architecture.

```docker
ARG TARGETARCH
ADD --chmod=755 https://github.com/articulate/docker-bootstrap/releases/latest/download/docker-bootstrap_linux_${TARGETARCH} /entrypoint

ENTRYPOINT [ "/entrypoint" ]
```

## Usage

To load values from Consul's KV store, you will need to set `CONSUL_ADDR`. It
will load keys from the following paths, using the basename as the variable name:

* `global/env_vars/*`
* `global/${SERVICE_ENV}/env_vars/*`
* `services/${SERVICE_NAME}/env_vars/*`
* `services/${SERVICE_NAME}/${SERVICE_ENV}/env_vars/*`

For example, `consul kv put services/foo/env_vars/API_SERVICE_URI https://api.priv/v1`
will load an environment variable `API_SERVICE_URI=https://api.priv/v1`.

Any environment variables set previous to calling the script, will not change.
Paths later in the list will overwrite any previous values. For example,
`global/env_vars/FOO` will be overwritten by `service/my-service/env_vars/FOO`.

To load values from Vault, you will need to set `VAULT_ADDR` and authenticate with
Vault (see below). Values from vault will use the `value` key as the variable value.
Values are read from the following paths:

* `secret/global/env_vars/*` (in `stage` or `prod`)
* `secret/global/${SERVICE_ENV}/env_vars/*`
* `secret/services/${SERVICE_NAME}/env_vars/*` (in `stage` or `prod`)
* `secret/services/${SERVICE_NAME}/${SERVICE_ENV}/env_vars/*`

For example, `vault write secret/services/foo/env_vars/API_KEY value=secretkey` will load
an environment variable `API_KEY=secretkey`. Values from Vault will overwrite
Consul values, but follow the same rules otherwise.

<details>
<summary>Vault Authentication</summary>

You can authenticate with Vault in one of the following ways:

* Set `VAULT_TOKEN`
* If running on Kubernetes, use the Kubernetes auth method in Vault
* If running on AWS ECS or Lambda, use the AWS IAM auth method
  * If Vault role does not match IAM role, set with `VAULT_ROLE`

</details>

### Environment Variables

If you want to ensure some environment variables exist before running your command,
you can include a JSON file called `service.json` in the working directory. The
entrypoint will parse this file and check that the configured environment variables
exist and are not empty.

```json
{
  "dependencies": {
    "env_vars": {
      "required": [
        "FOO",
        "BAR"
      ],
      "optional": [
        "BAZ"
      ]
    }
  }
}
```

If any optional environment variables are missing, it will log that, but continue
to run.

If any required environment variables are missing, it will log that and then exit
with an exit code of 4.

## Development

You'll need to install the following:

* Go
* [golangci-lint](https://golangci-lint.run/) (`brew install golangci-lint`)
* [pre-commit](https://pre-commit.com/) (`brew install pre-commit`)
* [GoReleaser](https://goreleaser.com/) (_optional_)

Setup the build environment with `make init`. Run tests with `make test` and lint
code with `make lint`.

When committing, you'll need to follow the [Conventional Commits](https://www.conventionalcommits.org)
format. You can install a tool like [git-cz](https://github.com/commitizen/cz-cli#conventional-commit-messages-as-a-global-utility)
or [commitizen](https://github.com/commitizen-tools/commitizen#installation).

## Creating a Release

To create a release, create a tag that follows [semver](https://semver.org/). A
GitHub Action workflow will take care of creating the release.
