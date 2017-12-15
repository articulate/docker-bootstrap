#!/bin/bash -e

AWS_REGION="${AWS_REGION:-us-east-1}"

if [ "${APP_ENV}" == "dev" ]
then
  # we use `source` here because we want dev-entrypoint.sh to be executed in the context
  # of this script, and not as a new process
  source /dev-entrypoint.sh
else
  MISBEHAVING_NOTICE="may be misbehaving. In a perfect world, our monitoring detected this problem and Platform Engineering was alerted... but just in case, please let us know."

  if [ ${CONSUL_ADDR} ]
  then
    if consul-template -consul-addr=$CONSUL_ADDR -template=/consul-template/${APP_ENV}/export-consul.ctmpl:/tmp/export-consul.sh -once -max-stale=0
    then
      source /tmp/export-consul.sh
    else
      echo "Consul $MISBEHAVING_NOTICE"
      exit 1
    fi
  else
    echo "CONSUL_ADDR are not set skipping Consul exports"
  fi

  if [ "${ENCRYPTED_VAULT_TOKEN}" ] && [ ! "${VAULT_TOKEN}" ]
  then
    export VAULT_TOKEN=$(echo $ENCRYPTED_VAULT_TOKEN | base64 -d | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 -d)
  fi

  if [ ${VAULT_TOKEN} ] && [ ${CONSUL_ADDR} ] && [ ${VAULT_ADDR} ]
  then
    if consul-template -consul-addr=$CONSUL_ADDR -vault-addr=$VAULT_ADDR -template=/consul-template/${APP_ENV}/export-vault.ctmpl:/tmp/export-vault.sh -once -max-stale=0
    then
      source /tmp/export-vault.sh
    else
      echo "Vault $MISBEHAVING_NOTICE"
      exit 1
    fi
  else
    echo "VAULT_TOKEN or VAULT_ADDR are not set, skipping Vault exports"
  fi

  rm -f /tmp/export-vault.sh
  rm -f /tmp/export-consul.sh
fi

exec "$@"
