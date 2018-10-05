#!/bin/bash

set -e

: "${SYSTEM_DOMAIN:?}"

bosh -n -d cf deploy cf-deployment/cf-deployment.yml --vars-store creds.yml -v system_domain=system.gcp.seankeery.com -o cf-deployment/operations/use-compiled-releases.yml -o cf-deployment/operations/scale-to-one-az.yml -o local.yml 
