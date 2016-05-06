#!/bin/bash
if [ ${VAULT_TOKEN+x} ]
then
  consul-template -template=/exports.ctmpl:exports.sh -ssl-verify=false -once
  source exports.sh
else
  echo "VAULT_TOKEN is not set skipping exports"
fi
exec "$@"
