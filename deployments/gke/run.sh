#!/bin/bash

set -e

bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  --state=state.json \
  --vars-store=creds.yml \
  -o ../../bosh-deployment/k8s/cpi.yml \
  -o ../../bosh-deployment/k8s/gcp.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../../manifests/dev.yml \
  -v director_name=kube-gke \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=35.224.114.62 \
  --var-file kube_config=<(cat ./kubeconfig) \
  -v kubernetes_cpi_path=/tmp/kube-cpi \
  --var-file gcr_password=gcr-password.json \
  -v gcr_pull_secret_name=regsecret \
  -v kube_api=104.198.249.174 \
  -o ../generic/local.yml
