#!/bin/bash
if [ ${VAULT_TOKEN+x} ]
then
  consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:/tmp/exports.sh -once
  exec "source /tmp/exports.sh && $*"
else
  echo "VAULT_TOKEN is not set skipping exports"
  exec "$@"
fi
