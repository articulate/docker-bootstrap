MISBEHAVING_NOTICE="may be misbehaving. Please ensure you have skipper and private-resources running properly."

if consul-template -consul-addr=consul-tugboat.dev.articulate.zone -template=/consul-template/${SERVICE_ENV}/export-consul.ctmpl:/tmp/export-consul.sh -once -max-stale=0
then
  source /tmp/export-consul.sh
else
  echo "Consul tugboat $MISBEHAVING_NOTICE"
  exit 1
fi

if consul-template -consul-addr=consul-priv.dev.articulate.zone -template=/consul-template/${SERVICE_ENV}/export-consul.ctmpl:/tmp/export-consul.sh -once -max-stale=0
then
  source /tmp/export-consul.sh
else
  echo "Consul stage $MISBEHAVING_NOTICE"
  exit 1
fi

if [ "${ENCRYPTED_VAULT_TOKEN}" ] && [ ! "${VAULT_TOKEN}" ]
then
  export VAULT_TOKEN=$(echo $ENCRYPTED_VAULT_TOKEN | base64 -d | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region $AWS_REGION | base64 -d)
fi

if [ ${VAULT_TOKEN} ] && [ ${CONSUL_ADDR} ] && [ ${VAULT_ADDR} ]
then
  if consul-template -consul-addr=consul-priv.dev.articulate.zone -vault-addr=http://vault-priv.dev.articulate.zone -template=/consul-template/${SERVICE_ENV}/export-vault.ctmpl:/tmp/export-vault.sh -once -max-stale=0
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
