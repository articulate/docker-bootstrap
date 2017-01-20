#!/bin/bash

AWS_REGION="${AWS_REGION:-us-east-1}"

if [ "${ENCRYPTED_VAULT_TOKEN}" ] && [ ! "${VAULT_TOKEN}" ]
then
  export VAULT_TOKEN=$(echo $ENCRYPTED_VAULT_TOKEN | base64 --decode | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 --decode)
fi

if [ ${CONSUL_ADDR} ]
then
  if consul-template -consul=$CONSUL_ADDR -template=/export-consul.ctmpl:/tmp/export-consul.sh -once -max-stale=0
  then
    source /tmp/export-consul.sh
  else
    echo "======== Consul may be misbehaving. If you are seeing this in prod, Engineering Ops have been alerted. ========"
    exit 1
  fi
else
  echo "CONSUL_ADDR are not set skipping Consul exports"
fi

if [ ${VAULT_TOKEN} ] && [ ${CONSUL_ADDR} ] && [ ${VAULT_ADDR} ] 
then
  if consul-template -consul=$CONSUL_ADDR -template=/export-vault.ctmpl:/tmp/export-vault.sh -once -max-stale=0
  then
    source /tmp/export-vault.sh
  else
    echo "======== Vault may be misbehaving. If you are seeing this in prod, Engineering Ops have been alerted. ========"
    exit 1
  fi
else
  echo "VAULT_TOKEN or VAULT_ADDR are not set, skipping Vault exports"
fi

exec "$@"
