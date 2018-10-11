#!/bin/bash

set -e

echo "-----> `date`: Deploy Director"
bosh delete-env ./bosh-deployment/bosh.yml \
  --state=state.json \
  --vars-store=creds.yml \
  -o ../../bosh-deployment/k8s/cpi.yml \
  -o ../../bosh-deployment/k8s/gcp.yml \
  -o ./bosh-deployment/jumpbox-user.yml \
  -o ../../manifests/dev.yml \
  -v director_name=kube-gke \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=${BOSH_RUN_LB_IP} \
  --var-file kube_config=<(cat ./kubeconfig) \
  -v kubernetes_cpi_path=/tmp/kube-cpi \
  --var-file gcr_password=gcr-password.json \
  -v gcr_pull_secret_name=regsecret \
  -v kube_api=${BOSH_RUN_KUBE_API_IP} \
  -o ../generic/local.yml
