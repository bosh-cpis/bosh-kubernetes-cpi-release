#!/bin/bash

set -e

: "${SYSTEM_DOMAIN:?}"

bosh -n -d cf deploy cf-deployment/cf-deployment.yml \
  --vars-store creds.yml \
  -v system_domain="${SYSTEM_DOMAIN}" \
  -o cf-deployment/operations/use-compiled-releases.yml \
  -o cf-deployment/operations/experimental/use-bosh-dns.yml \
  -o cf-deployment/operations/experimental/skip-consul-cell-registrations.yml \
  -o cf-deployment/operations/experimental/skip-consul-locks.yml \
  -o cf-deployment/operations/experimental/disable-consul.yml \
  -o cf-deployment/operations/experimental/use-grootfs.yml \
  -o local.yml
