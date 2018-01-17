#!/bin/bash

kubectl create -f ns.yml

ginkgo -r src/github.com/cppforlife/bosh-kubernetes-cpi/integration/
