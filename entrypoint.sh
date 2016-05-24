#!/bin/bash
if [ ${VAULT_TOKEN} ]
then
  consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:/tmp/exports.sh -once
  source /tmp/exports.sh
else
  echo "VAULT_TOKEN is not set skipping exports"
fi
exec "$@"
