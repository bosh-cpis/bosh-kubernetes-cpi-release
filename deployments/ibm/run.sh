#!/bin/bash

set -e

bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  --state=state.json \
  --vars-store=creds.yml \
  -o ../../bosh-deployment/k8s/cpi.yml \
  -o ../../bosh-deployment/k8s/ibm.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../../manifests/dev.yml \
  -v director_name=kube-ibm \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=159.122.196.34 \
  --var-file kube_config=<(cat ./kubeconfig) \
  -v kubernetes_cpi_path=/tmp/kube-cpi \
  --var-file gcr_password=gcr-password.json \
  -v gcr_pull_secret_name=regsecret \
  -v kube_api=159.122.242.78 \
  -v kube_port=12345 \
  -o ../generic/local.yml \
  -o local.yml
