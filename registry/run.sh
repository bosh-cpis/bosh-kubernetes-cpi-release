#!/bin/bash

set -e -u

eval $(minikube docker-env)

set -x

rm -rf certs && mkdir certs

bosh int certs.yml --vars-store certs/creds.yml \
  --var-file docker_ca.private_key=$DOCKER_CERT_PATH/ca-key.pem \
  --var-file docker_ca.certificate=$DOCKER_CERT_PATH/ca.pem \
  -v internal_ip=$(minikube ip)

bosh int certs/creds.yml --path /registry_tls/private_key > certs/domain.key
bosh int certs/creds.yml --path /registry_tls/certificate > certs/domain.crt

docker stop registry || true
docker rm registry || true

docker run -d \
  --restart=always \
  --name registry \
  -v `pwd`/certs:/certs \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/domain.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/domain.key \
  -p 443:443 \
  registry:2
