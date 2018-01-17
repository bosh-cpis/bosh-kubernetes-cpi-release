#!/bin/bash

set -e

bosh create-env --recreate ~/workspace/bosh-deployment/bosh.yml \
  --state=state.json \
  --vars-store=creds.yml \
  -o ../../bosh-deployment/k8s/cpi.yml \
  -o ../../bosh-deployment/k8s/aks.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../../manifests/dev.yml \
  -v director_name=kube-aks \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=13.92.138.87 \
  --var-file kube_config=<(cat ./kubeconfig) \
  -v kubernetes_cpi_path=/tmp/kube-cpi \
  --var-file gcr_password=gcr-password.json \
  -v gcr_pull_secret_name=regsecret \
  -v kube_api=myk8sclust-myresourcegroup-3c39a0-44371e09.hcp.eastus.azmk8s.io \
  -o ../generic/local.yml \
  -o local.yml
