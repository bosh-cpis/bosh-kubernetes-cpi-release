#!/bin/bash

set -e

echo "-----> Check minikube is operational"
minikube ip

eval $(minikube docker-env)

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
  -v kubernetes_cpi_path=/tmp/kube-cpi \
  -v internal_ip=$(minikube ip) \
  -v docker_host=$DOCKER_HOST \
  --var-file docker_tls.ca=$DOCKER_CERT_PATH/ca.pem \
  --var-file docker_tls.certificate=$DOCKER_CERT_PATH/cert.pem \
  --var-file docker_tls.private_key=$DOCKER_CERT_PATH/key.pem \
  -o ../generic/local.yml
