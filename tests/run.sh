#!/bin/bash

set -e # -x

cpi_path=$PWD/cpi
minikube_ip=$(minikube ip)

rm -f creds.yml rc-creds.yml

echo "-----> `date`: Create dev release"
bosh create-release --force --dir ./../ --tarball $cpi_path

echo "-----> `date`: Create kube namespace"
[ "x$(kubectl config current-context)" == "xminikube" ] || exit 1;
kubectl create -f ../deployments/generic/ns.yml || true

echo "-----> `date`: Create env"
bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  -o ../bosh-deployment/k8s/cpi.yml \
  -o ../bosh-deployment/k8s/minikube.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../manifests/dev.yml \
  --state=state.json \
  --vars-store=creds.yml \
  --var-file kube_config=<(cat ~/.kube/config) \
  -v kubernetes_cpi_path=$cpi_path \
  -v director_name=k8s \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=$minikube_ip

export BOSH_ENVIRONMENT=https://$minikube_ip:32001 # todo director port
export BOSH_CA_CERT="$(bosh int creds.yml --path /director_ssl/ca)"
export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET="$(bosh int creds.yml --path /admin_password)"

echo "-----> `date`: Update cloud config"
bosh -n update-cloud-config ../bosh-deployment/k8s/cloud-config.yml

echo "-----> `date`: Update runtime config"
bosh -n update-runtime-config ~/workspace/bosh-deployment/runtime-configs/dns.yml \
  --vars-store rc-creds.yml

echo "-----> `date`: Upload stemcell"
bosh -n upload-stemcell https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3468.17 \
  --sha1 1dad6d85d6e132810439daba7ca05694cec208ab

echo "-----> `date`: Create env second time to test persistent disk attachment"
bosh create-env ~/workspace/bosh-deployment/bosh.yml \
  -o ../bosh-deployment/k8s/cpi.yml \
  -o ../bosh-deployment/k8s/minikube.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../manifests/dev.yml \
  --state=state.json \
  --vars-store=creds.yml \
  --var-file kube_config=<(cat ~/.kube/config) \
  -v kubernetes_cpi_path=$cpi_path \
  -v director_name=k8s \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=$minikube_ip \
  --recreate

echo "-----> `date`: Delete previous deployment"
bosh -n -d zookeeper delete-deployment --force

echo "-----> `date`: Deploy"
bosh -n -d zookeeper deploy <(wget -O- https://raw.githubusercontent.com/cppforlife/zookeeper-release/master/manifests/zookeeper.yml)

echo "-----> `date`: Exercise deployment"
bosh -n -d zookeeper run-errand smoke-tests

echo "-----> `date`: Restart deployment"
bosh -n -d zookeeper restart

echo "-----> `date`: Report any problems"
bosh -n -d zookeeper cck --report

echo "-----> `date`: Delete random VM"
bosh -n -d zookeeper delete-vm `bosh -d zookeeper vms|sort|cut -f5|head -1`

echo "-----> `date`: Fix deleted VM"
bosh -n -d zookeeper cck --auto

echo "-----> `date`: Delete deployment"
bosh -n -d zookeeper delete-deployment

echo "-----> `date`: Clean up disks, etc."
bosh -n -d zookeeper clean-up --all

echo "-----> `date`: Deleting env"
bosh delete-env ~/workspace/bosh-deployment/bosh.yml \
  -o ../bosh-deployment/k8s/cpi.yml \
  -o ../bosh-deployment/k8s/minikube.yml \
  -o ~/workspace/bosh-deployment/jumpbox-user.yml \
  -o ../manifests/dev.yml \
  --state=state.json \
  --vars-store=creds.yml \
  --var-file kube_config=<(cat ~/.kube/config) \
  -v kubernetes_cpi_path=$cpi_path \
  -v director_name=k8s \
  -v internal_cidr="unused" \
  -v internal_gw="unused" \
  -v internal_ip=$minikube_ip

echo "-----> `date`: Done"
