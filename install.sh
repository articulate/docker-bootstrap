#!/bin/bash

apt-get update
apt-get install -y unzip git sudo python-dev jq wget
rm -rf /var/lib/apt/lists/*

git clone https://github.com/articulate/docker-consul-template-bootstrap.git
wget https://releases.hashicorp.com/consul-template/0.18.0-rc1/consul-template_0.18.0-rc1_linux_amd64.zip
unzip -d /usr/local/bin consul-template_0.18.0-rc1_linux_amd64.zip

wget "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip"
unzip awscli-bundle.zip
sudo ./awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws

mv ./docker-consul-template-bootstrap/entrypoint.sh /entrypoint.sh
mv ./docker-consul-template-bootstrap/export-consul.ctmpl /export-consul.ctmpl
mv ./docker-consul-template-bootstrap/export-vault.ctmpl /export-vault.ctmpl
