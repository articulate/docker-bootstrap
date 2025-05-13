#!/bin/sh
set -e

AWSCLI_VERSION="${AWSCLI_VERSION:-2.27.13}"

missing=""
command -v curl >/dev/null || missing="${missing} curl"
command -v unzip >/dev/null || missing="${missing} unzip"
test -d /etc/ssl/certs || missing="${missing} ca-certificates"

apt-get update -qq
# shellcheck disable=SC2086
apt-get install --yes --no-install-recommends $missing less groff

curl -s "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m)-${AWSCLI_VERSION}.zip" -o /tmp/awscliv2.zip
unzip -q -d /tmp /tmp/awscliv2.zip
/tmp/aws/install "$@"
rm -rf /tmp/aws /tmp/awscliv2 /tmp/awscliv2.zip

# shellcheck disable=SC2086
test -z "$missing" || apt-get remove --yes $missing
apt-get clean
apt-get autoclean
apt-get autoremove --yes --purge
rm -rf /var/lib/apt/lists /var/cache/apt/archives /usr/share/doc /root/.cache/
