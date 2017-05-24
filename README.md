# Docker Consul Template Bootstrap

This gets pulled in and sets up consul-template binaries and installs the entrypoint script and the exports.ctmpl file.

## Development Usage

In order to test this locally you will need to edit the `docker-compose.override.yml` and add the following:

```
environment:
    APP_NAME: "<your app name>"
    APP_ENV: "<dev||prod||stage>"
    VAULT_ADDR: "https://myarticulatetest.localtunnel.me"
    VAULT_TOKEN: "<the root vault token from your test vault server instance>"
```

You can run vault locally with `vault server -dev`. This command will output the VAULT_TOKEN you need and listen on port 8200. I use [localtunnel](https://localtunnel.me) to grab a url to use as the VAULT_ADDR.
`lt --port 8200 -s myarticulatetest`
