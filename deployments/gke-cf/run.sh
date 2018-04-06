#!/bin/bash

set -e

cf_lb_ip=$(kubectl -n bosh get svc cf-ingress -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

bosh -n -d cf deploy ~/workspace/cf-deployment/cf-deployment.yml \
  --vars-store cf-creds.yml \
  -v system_domain=${cf_lb_ip}.sslip.io \
  -o ~/workspace/cf-deployment/operations/use-compiled-releases.yml \
  -o ~/workspace/cf-deployment/operations/experimental/use-bosh-dns.yml \
  -o ~/workspace/cf-deployment/operations/experimental/skip-consul-cell-registrations.yml \
  -o ~/workspace/cf-deployment/operations/experimental/skip-consul-locks.yml \
  -o ~/workspace/cf-deployment/operations/experimental/disable-consul.yml \
  -o ~/workspace/cf-deployment/operations/experimental/use-grootfs.yml \
  -o ../gke-cf/local.yml
