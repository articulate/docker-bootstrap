#!/bin/bash

AWS_REGION="${AWS_REGION:-us-east-1}"

if [ "${ENCRYPTED_VAULT_TOKEN}" ] && [ ! "${VAULT_TOKEN}" ]
then
  export VAULT_TOKEN=$(echo $ENCRYPTED_VAULT_TOKEN | base64 --decode | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 --decode)
fi

if [ ${CONSUL_ADDR} ]
then
  if consul-template -consul=$CONSUL_ADDR -template=/consul-exports.ctmpl:/tmp/consul-exports.sh -once -max-stale=0
  then
    source /tmp/consul-exports.sh
  else
    echo "======== Consul may be misbehaving. If you are seeing this in prod, Engineering Ops have been alerted. ========"
    exit 1
  fi
else
  echo "CONSUL_ADDR are not set skipping Consul exports"
fi
exec "$@"


if [ ${VAULT_TOKEN} ] && [ ${CONSUL_ADDR} ] && [ ${VAULT_ADDR} ] 
then
  if consul-template -consul=$CONSUL_ADDR -template=/vault-exports.ctmpl:/tmp/vault-exports.sh -once -max-stale=0
  then
    source /tmp/vault-exports.sh
  else
    echo "======== Vault may be misbehaving. If you are seeing this in prod, Engineering Ops have been alerted. ========"
    exit 1
  fi
else
  echo "VAULT_TOKEN or VAULT_ADDR are not set, skipping Vault exports"
fi
exec "$@"
