#!/bin/bash

if [ -f ".app.json" ] && [ "${APP_ENV}" == "stage" ]
then
  export VAULT_TOKEN=$(jq -r '.vault_token' .app.json | base64 --decode | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 --decode)
fi

if [ ${VAULT_TOKEN} ]
then
  if consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:/tmp/exports.sh -once -max-stale=0
  then
    source /tmp/exports.sh
  else
    echo "======== CONSUL OR VAULT ARE HAVING ISSUES, WEB OPS HAVE BEEN ALERTED ========"
    exit 1
  fi
else
  echo "VAULT_TOKEN is not set skipping exports"
fi
exec "$@"
