#!/bin/bash

set -e

bosh_lb_ip=35.224.114.62
kube_api_ip=104.198.249.174

echo "-----> `date`: Create dev release"
cpi_path=/tmp/kube-cpi
bosh create-release --force --dir ./../../ --tarball $cpi_path

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
  -v internal_ip=${bosh_lb_ip} \
  --var-file kube_config=<(cat ./kubeconfig) \
  -v kubernetes_cpi_path=/tmp/kube-cpi \
  --var-file gcr_password=gcr-password.json \
  -v gcr_pull_secret_name=regsecret \
  -v kube_api=${kube_api_ip} \
  -o ../generic/local.yml
