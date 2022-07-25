# Docker Consul Template Bootstrap

This gets pulled in and sets up consul-template binaries and installs the entrypoint script and the exports.ctmpl file.

## Circumventing Docker Caching
Because of this dynamic nature, we will need to update the install.sh slightly if this repository is modified. See the `CACHE VERSION` at the top of install.sh. 

## Development Usage

In order to test this locally you will need to edit the `docker-compose.override.yml` and add the following:

```yaml
environment:
    SERVICE_NAME: "your-service"
    SERVICE_ENV: "dev|stage|prod|peer"
    VAULT_ADDR: "https://myarticulatetest.localtunnel.me"
    VAULT_TOKEN: "your-token"
```

You can run vault locally with `vault server -dev`. This command will output the VAULT_TOKEN you need and listen on port 8200. I use [localtunnel](https://localtunnel.me) to grab a url to use as the VAULT_ADDR.
`lt --port 8200 -s myarticulatetest`

## Test Suite

The test suite is written in rspec and creates a series of containers (both vault & consul) and runs a series of tests against
those. The module used within rspec is `https://github.com/zuazo/dockerspec` and uses serverspec behind the scenes.

The tests included run through the normal cascade pattern of Global -> Product -> Service and at the end provides output.
It is not uncommon for these comprehensive tests to take ~10 mins.

To kick off the tests:

1. Create a branch and push that branch to GitHub.
2. Run something similar to `cp docker-compose.override.example.yml docker-compose.override.yml`

Then run:

`docker-compose run app`

If your changes are not commited and pushed it will not be picked up
when the test images build (rspec pulls in details from the git branch to create the tests).

There may be a chance that the containers created by the test suite contain cached content.
If needed you can run:

```bash
docker-compose down
docker rmi --force docker-consul-template-bootstrap_app
docker rmi --force consul_template_bootstrap_alpine
docker rmi --force consul_template_bootstrap_centos
docker rmi --force consul_template_bootstrap_debian
```
