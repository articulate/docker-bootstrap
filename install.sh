#!/bin/bash
set -eo pipefail

# Due to docker's layer caching, you may need to update this file in a way to force docker
# to skip the layer cache and re-run this install the next time it builds the image.
# Simply edit the date here: 2021-06-11.1

CONSUL_TEMPLATE_BOOTSTRAP_REF="${1:-master}"

if command -v apt-get; then
  apt-get update
  apt-get -y install --no-install-recommends unzip sudo jq wget curl ca-certificates
  apt-get clean && apt-get autoclean && apt-get -y autoremove --purge
  rm -rf /var/lib/apt/lists/* /usr/share/doc /root/.cache/

  # AWS CLI
  curl -s "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o /tmp/awscliv2.zip
  unzip -d /tmp /tmp/awscliv2.zip
  /tmp/aws/install
  rm -rf /tmp/aws /tmp/awscliv2
elif command -v yum; then
  grep "Amazon Linux" /etc/os-release &>/dev/null || yum -y install epel-release
  yum -y update
  yum -y install unzip jq sudo wget curl which
  yum clean all

  # AWS CLI
  curl -s "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o /tmp/awscliv2.zip
  unzip -d /tmp /tmp/awscliv2.zip
  /tmp/aws/install
  rm -rf /tmp/aws /tmp/awscliv2
elif command -v apk; then
  apk add --no-cache --update unzip sudo python3 jq wget ca-certificates curl which py3-pip
  update-ca-certificates
  rm -rf /var/cache/apk/*

  # Use the Python version of AWS CLI since they don't provide a musl compatible build
  pip3 --no-cache-dir install awscli
else
  echo "Existing package manager is not supported"
  exit 1
fi

# Install Consul template
CONSUL_TEMPLATE_VERSION=0.25.1
wget -q -O /tmp/consul-template.zip "https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip"
unzip -d /usr/local/bin /tmp/consul-template.zip
rm /tmp/consul-template.zip

# Install Vault CLI
VAULT_VERSION=1.5.4
wget -q -O /tmp/vault.zip "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip"
unzip -d /tmp /tmp/vault.zip
mv /tmp/vault /usr/local/bin/vault
chmod +x /usr/local/bin/vault
rm -rf /tmp/vault*

# Install consul-bootstrap
wget -q -O /tmp/docker-consul-template-bootstrap.zip "https://github.com/articulate/docker-consul-template-bootstrap/archive/${CONSUL_TEMPLATE_BOOTSTRAP_REF}.zip"
unzip -d /tmp /tmp/docker-consul-template-bootstrap.zip
mv "/tmp/docker-consul-template-bootstrap-${CONSUL_TEMPLATE_BOOTSTRAP_REF}/" /consul-template/
mv /consul-template/entrypoint.sh /entrypoint.sh
rm /tmp/docker-consul-template-bootstrap.zip

for package in wget jq curl which aws consul-template vault; do
  if ! command -v "$package"; then
    echo "$package is not installed"
    exit 1
  fi
done
