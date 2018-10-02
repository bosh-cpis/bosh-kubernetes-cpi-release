## Stemcell

CPI is configured to use GCR with heavy warden stemcells.

## Director

Steps:

- `cd ./deployments/gke/`
- create GKE cluster
  - give it a name
  - set location to be zonal (avoids issues where Nodes and Persistent disks are created in different Zones)
  - select 2 CPUs per node
  - set Node image to Ubuntu (Container-Optimized OS (cos) does not support xfs filesystem used by Garden)
  - select size to be 3 nodes
  - accept default service account (in advanced section;)
    - if cluster creation fails, try using different service account
- grab cluster credentials
  - copy username/password/ca-certificate from cluster info page
  - copy cluster IP address
- create `./kubeconfig` from `./kubeconfig.example`
  - replace `((...))` with above values
- create `./kubeconfigca.pem` and paste cluster CA certificate
  - make sure to have CA properly encoded in PEM format
- `source .kube.envrc` to load kubeconfig
- `kubectl create -f ../generic/ns.yml`
- `kubectl create -f ../generic/lb.yml`
- set `export BOSH_RUN_LB_IP=` to the External IP of the new LoadBalancer
  - find LB IP via `kubectl -n bosh get svc` (may have to wait a min)
- set `export BOSH_RUN_KUBE_API_IP=` to cluster API IP
- create GCR credential
  - create new service account with "storage admin" role
    - https://cloud.google.com/container-registry/docs/access-control
  - create json key for the service account and rename to `./gcr-password.json`
- create secret for pulling images
  - `kubectl create secret docker-registry regsecret --docker-server=https://gcr.io --docker-username=_json_key --docker-password="$(cat gcr-password.json)" --docker-email foo`
- allow CPI to access different APIs
  - `kubectl create -f ../generic/cpi-rbac.yml`
- execute `./run.sh`
- run `source .bosh.envrc` and `bosh env` to verify bosh is accessible

Nothing special.

## Kafka & Zookeeper

Steps:

- `bosh update-cloud-config ../../bosh-deployment/k8s/cloud-config.yml`
- `bosh update-runtime-config ./bosh-deployment/runtime-configs/dns.yml --vars-store dns-creds.yml`
- `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3541.12 --sha1 14bd6dd50d3caa913af97846eab39e5075b240d7`
- Install zookeeper
  - `git clone https://github.com/cppforlife/zookeeper-release`
  - `bosh -d zookeeper deploy zookeeper-release/manifests/zookeeper.yml -o zookeeper-release/manifests/enable-dns.yml`
  - `bosh -d zookeeper run-errand status`
  - `bosh -d zookeeper run-errand smoke-tests`
- Install kafka
  - `git clone https://github.com/cppforlife/kafka-release`
  - `bosh -d kafka deploy kafka-release/manifests/example.yml -o kafka-release/manifests/enable-dns.yml`
  - `bosh -d kafka run-errand smoke-tests`

Nothing special.

## CF

See `deployments/gke-cf/`

- `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3541.12 --sha1 14bd6dd50d3caa913af97846eab39e5075b240d7`
- `cd ../deployments/gke-cf`
- `bosh update-cloud-config ./cc.yml`
- `kubectl create -f ../generic/lb-cf.yml`
- Run: `kubectl -n bosh get svc` to get the External IP of the new LoadBalancer.
- Create a wildcard DNS record of type A on GCP under Network Services with the
  IP from the previous step.
- Using the domain from the DNS record, run the script

```
SYSTEM_DOMAIN=YOUR-SYSTEM-DOMAIN ./run.sh
```
