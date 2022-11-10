# Docker Consul Template Bootstrap

Load values from Consul and Vault as environment variables.

## Installing

Download the [latest release](https://github.com/articulate/docker-consul-template-bootstrap/releases/latest),
add it to your image, and set it as your `ENTRYPOINT`.

If you are using Buildkit you can use the `TARGETARCH` arg to `ADD` the correct
architecture.

```docker
ARG TARGETARCH
ADD --chmod=755 https://github.com/articulate/docker-consul-template-bootstrap/releases/latest/download/docker-consul-template-bootstrap_linux_${TARGETARCH} /entrypoint

ENTRYPOINT [ "/entrypoint" ]
```

## Usage

To load a environment variables, you'll need to set a `SERVICE_ENV` environment
variable to `prod`, `stage`, `dev`, or `peer`. The following paths get loaded as
environment variables, but some environments may change this. You can view exact
paths used in each template.

* `global/env_vars/*` (_Consul_)
* `services/${SERVICE_NAME}/env_vars/*` (_Consul_)
* `secrets/global/env_vars/*` (_Vault_ using the `value` key)
* `services/${SERVICE_NAME}/env_vars/*` (_Vault_ using the `value` key)

To load values from Consul, you'll need to make sure `CONSUL_ADDR` is accessible
from your Docker container.

To load values from Vault, you'll need to make sure both `CONSUL_ADDR` and `VAULT_ADDR`
are accessible. You'll also need to authenticate with Vault in one of the following
ways:

* Set `VAULT_TOKEN`
* Set `ENCRYPTED_VAULT_TOKEN` with a value encrypted by AWS KMS
  * You'll need to make sure the container has permissions for the default KMS key
* If running on Kubernetes, use the Kubernetes auth method in Vault
* If running on AWS ECS or Lambda, use the AWS IAM auth method
  * If Vault role does not match IAM role, set with `VAULT_ROLE`

## Development

You'll need to install the following:

* Go 1.19
* [golangci-lint](https://golangci-lint.run/) (`brew install golangci-lint`)
* [pre-commit](https://pre-commit.com/) (`brew install pre-commit`)
* [GoReleaser](https://goreleaser.com/) (_optional_)

Setup the build environment with `make init`. Run tests with `make test` and lint
code with `make lint`.

When committing, you'll need to follow the [Conventional Commits](https://www.conventionalcommits.org)
format. You can install a tool like [git-cz](https://github.com/commitizen/cz-cli#conventional-commit-messages-as-a-global-utility)
or [commitizen](https://github.com/commitizen-tools/commitizen#installation).

## Creating a Release

To create a release, create a tag that follows [semver](https://semver.org/) and
a GitHub Action workflow will take care of creating the release.
