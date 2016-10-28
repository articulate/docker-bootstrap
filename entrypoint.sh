#!/bin/bash
if [ ${VAULT_TOKEN} ]
then
  if consul-template -consul=$CONSUL_ADDR -template=/exports.ctmpl:/tmp/exports.sh -once
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
