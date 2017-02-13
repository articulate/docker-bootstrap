#!/bin/bash -e

if [ `command -v apt-get` ]; then
  apt-get update
  apt-get install -y unzip git sudo python-dev jq wget
  rm -rf /var/lib/apt/lists/*
elif [ `command -v yum` ]; then
  yum update
  yum -y install unzip git python-devel jq wget
  yum clean all
else
  echo "Existing package manager is not supported"
  exit 1
fi

export CONSUL_TEMPLATE_VERSION=0.18.1
wget -O consul-template.zip https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip
unzip -d /usr/local/bin consul-template.zip
rm consul-template.zip

wget "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip"
unzip awscli-bundle.zip
sudo ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws
rm awscli-bundle.zip

git clone https://github.com/articulate/docker-consul-template-bootstrap.git

mv ./docker-consul-template-bootstrap/entrypoint.sh /entrypoint.sh
mv ./docker-consul-template-bootstrap/export-consul.ctmpl /export-consul.ctmpl
mv ./docker-consul-template-bootstrap/export-vault.ctmpl /export-vault.ctmpl
