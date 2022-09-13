# Docker Consul Template Bootstrap

Load values from Consul and Vault as environment variables.

## Installing

Run `install.sh` in your Docker image. This will copy over _entrypoint.sh_, Consul
template files, and install dependencies needed to run the entrypoint. Then set
your `ENTRYPOINT`.

```docker
ADD https://raw.githubusercontent.com/articulate/docker-consul-template-bootstrap/master/install.sh /tmp/consul_template_install.sh

RUN bash /tmp/consul_template_install.sh && rm /tmp/consul_template_install.sh

ENTRYPOINT [ "/entrypoint.sh" ]
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

## Circumventing Docker Caching

With Docker cache, if you make any changes outside of `install.sh` (e.g. the `ctmpl`
files), you also need to update `install.sh`. Update the `CACHE_VERSION` to the
current datetime.

## Development Usage

To test this locally you will need to edit the `docker-compose.override.yml` and
add the following:

```yaml
environment:
    SERVICE_NAME: "your-service"
    SERVICE_ENV: "dev|stage|prod|peer"
    VAULT_ADDR: "https://myarticulatetest.localtunnel.me"
    VAULT_TOKEN: "your-token"
```

You can run vault locally with `vault server -dev`. This command will output the
VAULT_TOKEN you need and listen on port 8200. I use [localtunnel](https://localtunnel.me)
to grab a url to use as the VAULT_ADDR. `lt --port 8200 -s myarticulatetest`

## Test Suite

The test suite is written in rspec and creates a series of containers (both vault
& consul) and runs a series of tests against those. The module used within rspec
is `https://github.com/zuazo/dockerspec` and uses serverspec behind the scenes.

The tests included run through the normal cascade pattern of Global -> Product ->
Service and at the end provides output. These comprehensive tests could take 10
minutes.

To kick off the tests:

1. Create a branch and push that branch to GitHub.
2. Run something similar to `cp docker-compose.override.example.yml docker-compose.override.yml`

Then run:

`docker-compose run app`

If your changes are not committed and pushed it will not be picked up when the test
images build (rspec pulls in details from the git branch to create the tests).

There may be a chance that the containers created by the test suite contain cached
content. If needed you can run:

```bash
docker-compose down
docker rmi --force docker-consul-template-bootstrap_app
docker rmi --force consul_template_bootstrap_alpine
docker rmi --force consul_template_bootstrap_centos
docker rmi --force consul_template_bootstrap_debian
```
