#!/bin/bash

set -e

: "${SYSTEM_DOMAIN:?}"

bosh -n -d cf deploy cf-deployment/cf-deployment.yml   --vars-store creds.yml   -v system_domain=system.gcp.seankeery.com -o local.yml -o cf-deployment/operations/experimental/rootless-containers.yml 
