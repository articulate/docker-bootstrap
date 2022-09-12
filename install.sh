#!/usr/bin/env bash

# NOTE: Update if changes are made to this repo that need to be included in the images
# CACHE VERSION: 202209121400

set -eo pipefail

CONSUL_TEMPLATE_BOOTSTRAP_REF="${1:-master}"

if command -v apt-get; then
  apt-get update
  apt-get -y install --no-install-recommends unzip sudo jq wget curl ca-certificates
  apt-get clean && apt-get autoclean && apt-get -y autoremove --purge
  rm -rf /var/lib/apt/lists/* /usr/share/doc /root/.cache/
elif command -v yum; then
  grep "Amazon Linux" /etc/os-release &>/dev/null || yum -y install epel-release
  yum -y update
  yum -y install unzip jq sudo wget curl which
  yum clean all
  rm -rf /var/cache/yum
elif command -v apk; then
  apk add --no-cache --update unzip sudo python3 jq wget ca-certificates curl which py3-pip bash
  update-ca-certificates
  rm -rf /var/cache/apk/*

  # Use the Python version of AWS CLI since they don't provide a musl compatible build
  pip3 --no-cache-dir install awscli
  SKIP_AWS_INSTALL=1
else
  echo "Could not find a supported package manager (apt-get, yum, apk)."
  exit 1
fi

if [ -z "$SKIP_AWS_INSTALL" ]; then
  curl -s "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o /tmp/awscliv2.zip
  unzip -d /tmp /tmp/awscliv2.zip
  /tmp/aws/install
  rm -rf /tmp/aws /tmp/awscliv2 /tmp/awscliv2.zip
fi

arch="linux_amd64"
[ "$(uname -m)" == "aarch64" ] && arch="linux_arm64"

# Install Consul template
CONSUL_TEMPLATE_VERSION="${CONSUL_TEMPLATE_VERSION:-0.28.1}"
curl -s "https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_${arch}.zip" -o /tmp/consul-template.zip
unzip /tmp/consul-template.zip consul-template -d /usr/local/bin
rm /tmp/consul-template.zip

# Install Vault CLI
VAULT_VERSION="${VAULT_VERSION:-1.10.0}"
curl -s "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_${arch}.zip" -o /tmp/vault.zip
unzip /tmp/vault.zip vault -d /usr/local/bin
rm /tmp/vault.zip

# Install consul-bootstrap
curl -Ls "https://github.com/articulate/docker-consul-template-bootstrap/archive/${CONSUL_TEMPLATE_BOOTSTRAP_REF}.zip" -o /tmp/docker-consul-template-bootstrap.zip
unzip /tmp/docker-consul-template-bootstrap.zip -d /tmp
mkdir -p /consul-template/
mv "/tmp/docker-consul-template-bootstrap-${CONSUL_TEMPLATE_BOOTSTRAP_REF}"/{dev,peer,prod,stage} /consul-template/
mv "/tmp/docker-consul-template-bootstrap-${CONSUL_TEMPLATE_BOOTSTRAP_REF}/entrypoint.sh" /entrypoint.sh
rm -rf /tmp/docker-consul-template-bootstrap*

for package in wget jq curl which aws consul-template vault; do
  if ! command -v "$package"; then
    echo "$package is not installed"
    exit 1
  fi
done
