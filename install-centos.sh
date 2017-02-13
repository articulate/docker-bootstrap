#!/bin/bash

yum install -y unzip git \
  && yum clean all

git clone https://github.com/articulate/docker-consul-template-bootstrap.git
wget https://releases.hashicorp.com/consul-template/0.18.0-rc1/consul-template_0.18.0-rc1_linux_amd64.zip
unzip -d /usr/local/bin consul-template_0.18.0-rc1_linux_amd64.zip

mv ./docker-consul-template-bootstrap/entrypoint.sh /entrypoint.sh
mv ./docker-consul-template-bootstrap/exports.ctmpl /exports.ctmpl

