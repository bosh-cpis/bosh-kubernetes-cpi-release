#!/bin/bash

bosh -n -d cf deploy ~/workspace/cf-deployment/cf-deployment.yml \
  --vars-store creds.yml \
  -v system_domain=52.170.18.97.sslip.io \
  -o ~/workspace/cf-deployment/operations/use-compiled-releases.yml \
  -o ~/workspace/cf-deployment/operations/experimental/use-bosh-dns.yml \
  -o ~/workspace/cf-deployment/operations/experimental/skip-consul-cell-registrations.yml \
  -o ~/workspace/cf-deployment/operations/experimental/skip-consul-locks.yml \
  -o ~/workspace/cf-deployment/operations/experimental/disable-consul.yml \
  -o ~/workspace/cf-deployment/operations/experimental/use-grootfs.yml \
  -o local.yml
