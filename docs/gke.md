## Stemcell

CPI is configured to use GCR with heavy warden stemcells.

## Director

Steps:

- `cd ./deployments/gke/`
- create GKE cluster
  - give it a name
  - select 2 CPUs per node
  - select 3 nodes
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
- set `export BOSH_RUN_LB_IP=` to LB IP
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

Requirements:

- must select GKE Ubuntu image instead of COS
- must reboot nodes with `cgroup_enable=memory swapaccount=1` in `/boot/grub/grub.cfg`
  - we were not able to use update-grub utility
  - verify that `cat /proc/cmdline` contains above configuration
