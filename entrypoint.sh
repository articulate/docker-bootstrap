#!/bin/bash -e

AWS_REGION="${AWS_REGION:-us-east-1}"
MISBEHAVING_NOTICE="may be misbehaving. In a perfect world, our monitoring detected this problem and Engineering Ops were alerted... but just in case, please let us know."

if [ "${ENCRYPTED_VAULT_TOKEN}" ] && [ ! "${VAULT_TOKEN}" ]
then
  export VAULT_TOKEN=$(echo $ENCRYPTED_VAULT_TOKEN | base64 --decode | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 --decode)
fi

if [ ${CONSUL_ADDR} ]
then
  if consul-template -consul-addr=$CONSUL_ADDR -template=/export-consul.ctmpl:/tmp/export-consul.sh -once -max-stale=0
  then
    source /tmp/export-consul.sh
  else
    echo "Consul $MISBEHAVING_NOTICE"
    exit 1
  fi
else
  echo "CONSUL_ADDR are not set skipping Consul exports"
fi

if [ ${VAULT_TOKEN} ] && [ ${CONSUL_ADDR} ] && [ ${VAULT_ADDR} ]
then
  if consul-template -consul-addr=$CONSUL_ADDR -vault-addr=$VAULT_ADDR -template=/export-vault.ctmpl:/tmp/export-vault.sh -once -max-stale=0
  then
    source /tmp/export-vault.sh
  else
    echo "Vault $MISBEHAVING_NOTICE"
    exit 1
  fi
else
  echo "VAULT_TOKEN or VAULT_ADDR are not set, skipping Vault exports"
fi

exec "$@"
