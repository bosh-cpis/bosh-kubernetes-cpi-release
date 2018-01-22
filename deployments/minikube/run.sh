#!/bin/bash

set -e

echo "-----> `date`: Check minikube is operational"
minikube_ip=$(minikube ip)
eval $(minikube docker-env)

echo "-----> `date`: Create dev release"
cpi_path=/tmp/kube-cpi
bosh create-release --force --dir ./../../ --tarball $cpi_path

echo "-----> `date`: Deploying"
bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  --state=state.json \
  --vars-store=creds.yml \
  -o ../../bosh-deployment/k8s/cpi.yml \
  -o ../../bosh-deployment/k8s/minikube.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../../manifests/dev.yml \
  -v director_name=kube-minikube \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  --var-file kube_config=<(cat ~/.kube/config) \
  -v kubernetes_cpi_path=$cpi_path \
  -v internal_ip=$minikube_ip \
  -v docker_host=$DOCKER_HOST \
  --var-file docker_tls.ca=$DOCKER_CERT_PATH/ca.pem \
  --var-file docker_tls.certificate=$DOCKER_CERT_PATH/cert.pem \
  --var-file docker_tls.private_key=$DOCKER_CERT_PATH/key.pem \
  -o ../generic/local.yml
