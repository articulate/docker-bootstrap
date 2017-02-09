#!/bin/bash

AWS_REGION="${AWS_REGION:-us-east-1}"

if [ "${ENCRYPTED_VAULT_TOKEN}" ] && [ ! "${VAULT_TOKEN}" ]
then
  export VAULT_TOKEN=$(echo $ENCRYPTED_VAULT_TOKEN | base64 --decode | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 --decode)
fi

if [ ${VAULT_TOKEN} ] && [ ${CONSUL_ADDR} ] && [ ${VAULT_ADDR} ]
then
  if consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:/tmp/exports.sh -once -max-stale=0
  then
    source /tmp/exports.sh
  else
    echo "======== Consul or Vault are misbehaving. If you are seeing this in prod, Engineering Ops have been alerted. ========"
    exit 1
  fi
else
  echo "VAULT_TOKEN, VAULT_ADDR, or CONSUL_ADDR are not set skipping exports"
fi
exec "$@"
