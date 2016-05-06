#!/bin/bash

apt-get update
apt-get install -y unzip git
rm -rf /var/lib/apt/lists/*
git clone https://github.com/articulate/docker-consul-template-bootstrap.git
wget https://releases.hashicorp.com/consul-template/0.14.0/consul-template_0.14.0_linux_amd64.zip
unzip -d /usr/local/bin consul-template_0.14.0_linux_amd64.zip

mv ./docker-consul-template-bootstrap/entrypoint.sh /entrypoint.sh
mv ./docker-consul-template-bootstrap/exports.ctmpl /exports.ctmpl

