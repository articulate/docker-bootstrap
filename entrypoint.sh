#!/bin/bash
if [ ${VAULT_TOKEN} ]
then
  echo "VAULT_TOKEN is present generating /tmp/exports.sh"
  consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:/tmp/exports.sh -once
  source /tmp/exports.sh
else
  echo "VAULT_TOKEN is not set skipping exports"
fi
exec "$@"
