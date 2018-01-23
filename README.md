# Docker Consul Template Bootstrap

[![Build Status](https://travis-ci.org/articulate/docker-consul-template-bootstrap.svg?branch=master)](https://travis-ci.org/articulate/docker-consul-template-bootstrap)

This gets pulled in and sets up consul-template binaries and installs the entrypoint script and the exports.ctmpl file.

## Development Usage

In order to test this locally you will need to edit the `docker-compose.override.yml` and add the following:

```
environment:
    SERVICE_NAME: "your-service"
    SERVICE_ENV: "dev|stage|prod|peer"
    CONSUL_ADDR: "http://consul.articulate.zone"
    VAULT_ADDR: "https://myarticulatetest.localtunnel.me"
    VAULT_TOKEN: "your-token"
```

You can run vault locally with `vault server -dev`. This command will output the VAULT_TOKEN you need and listen on port 8200. I use [localtunnel](https://localtunnel.me) to grab a url to use as the VAULT_ADDR.
`lt --port 8200 -s myarticulatetest`

## Test Suite

To run the test suite, first make sure your changes are commited and
pushed to github.  Then run:

`docker-compose run app`

If your changes are not commited and pushed it will not be picked up
when the test images build.
