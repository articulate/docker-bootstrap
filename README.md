# Docker Consul Template Bootstrap

This gets pulled in and sets up consul-template binaries and installs the entrypoint script and the exports.ctmpl file.

## Using Vault

In order to use vault to pull in secrets as ENV variables you will need to include the following Atlas variables in your EB settings.

- The ENV variables `VAULT_ADDR` and `VAULT_TOKEN` will need to be defined in atlas variables and in the terraform module as Elasticbeanstalk environment settings. Here is an example of the settings for the terraform module. These will be used by consul-template to connect to vault.
 ```
setting {
  namespace = "aws:elasticbeanstalk:application:environment"
  name = "APP_NAME"
  value = "<your app name>"
  }
setting {
  namespace = "aws:elasticbeanstalk:application:environment"
  name = "VAULT_TOKEN"
  value = "${lookup(var.my_app_vault_token, var.env)}"
  }
setting {
  namespace = "aws:elasticbeanstalk:application:environment"
  name = "VAULT_ADDR"
  value = "${lookup(var.vault_addr, var.env)}"
  }
```
The app name, repo name and name in the vault path should all be the same.

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
