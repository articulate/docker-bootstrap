#!/bin/bash
if [ ${VAULT_TOKEN+x} ]
then
  consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:exports.sh -once
  source exports.sh
else
  echo "VAULT_TOKEN is not set skipping exports"
fi
exec "$@"
