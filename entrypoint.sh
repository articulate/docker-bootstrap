#!/usr/bin/env bash
set -e

AWS_REGION="${AWS_REGION:-us-east-1}"

# This will return everything before a - chararacter.
# "peer-something-thing" => "peer"
CT_SERVICE_ENV="${SERVICE_ENV%%-*}"

if [ -n "$CONSUL_ADDR" ]; then
  if consul-template -consul-addr="$CONSUL_ADDR" -template="/consul-template/${CT_SERVICE_ENV}/export-consul.ctmpl:/tmp/export-consul.sh" -once -max-stale=0; then
    # shellcheck disable=SC1091
    source /tmp/export-consul.sh
    rm -f /tmp/export-consul.sh
  else
    (>&2 echo "ERROR: Unable to export from Consul")
    exit 1
  fi
else
  (>&2 echo "WARN: CONSUL_ADDR is not set, skipping Consul exports")
fi

if [ -n "$VAULT_ADDR" ]; then
  if [ -z "$VAULT_TOKEN" ] && [ -f /var/run/secrets/kubernetes.io/serviceaccount/token ]; then
    KUBE_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
    VAULT_TOKEN="null"
    attempts=1
    while [ "${attempts}" -le 10 ]; do
      response=$(curl -s --show-error --request POST \
        --data '{"jwt": "'"$KUBE_TOKEN"'", "role": "'"$SERVICE_NAME"'"}' \
        "${VAULT_ADDR}/v1/auth/kubernetes/login")
      VAULT_TOKEN=$(echo "$response" | jq -r '.auth.client_token')
      [[ "$VAULT_TOKEN" != "null" ]] && break
      (>&2 echo "Attempt number ${attempts} to get vault token failed, retrying...")
      (>&2 echo "Vault response: $(echo "$response" | jq '.errors[]')")
      ((attempts+=1))
      sleep 3
    done

    if [[ "$VAULT_TOKEN" == "null" ]]; then
      (>&2 echo "ERROR: Unable to get vault token via kubernetes")
      exit 1
    fi

    export VAULT_TOKEN
  fi

  if [ -n "$ENCRYPTED_VAULT_TOKEN" ] && [ -z "$VAULT_TOKEN" ]; then
    VAULT_TOKEN=$(echo "$ENCRYPTED_VAULT_TOKEN" | base64 -d | aws kms decrypt --ciphertext-blob fileb:///dev/stdin --output text --query Plaintext --region "$AWS_REGION" | base64 -d)
    export VAULT_TOKEN
  fi

  if [ -z "$VAULT_TOKEN" ] && [ -n "$AWS_CONTAINER_CREDENTIALS_RELATIVE_URI" ]; then
    [ -n "$VAULT_ROLE" ] && role="role=${VAULT_ROLE}"
    VAULT_TOKEN=$(vault login -method=aws -token-only "$role")
    export VAULT_TOKEN
  fi

  if [ -z "$VAULT_TOKEN" ] && [ -n "$AWS_LAMBDA_FUNCTION_NAME" ]; then
    [ -n "$VAULT_ROLE" ] && role="role=${VAULT_ROLE}"
    VAULT_TOKEN=$(vault login -method=aws -token-only "$role")
    export VAULT_TOKEN
  fi

  if [ -n "$VAULT_TOKEN" ] && [ -n "$CONSUL_ADDR" ]; then
    if consul-template -consul-addr="$CONSUL_ADDR" -vault-addr="$VAULT_ADDR" -template="/consul-template/${CT_SERVICE_ENV}/export-vault.ctmpl:/tmp/export-vault.sh" -once -max-stale=0; then
      # shellcheck disable=SC1091
      source /tmp/export-vault.sh
      rm -f /tmp/export-vault.sh
    else
      (>&2 echo "ERROR: Unable to export from Vault")
      exit 1
    fi
  else
    (>&2 echo "WARN: VAULT_TOKEN and/or CONSUL_ADDR not set, skipping Vault exports")
  fi
else
  (>&2 echo "WARN: VAULT_ADDR is not set, skipping Vault exports")
fi

exec "$@"
