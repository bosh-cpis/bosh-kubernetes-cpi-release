# BOSH Kubernetes CPI

The BOSH Kubernetes CPI allows BOSH to manage deploy BOSH workloads such as CF onto Kubernetes clusters.

... example: minikube ...

... example: gke ...

## Use with Kube environments

- [GKE](docs/gke.md)
- [IBM](docs/ibm.md)
- [Minikube](docs/minikube.md)

## Development

- unit tests
  - `./src/github.com/cppforlife/bosh-kubernetes-cpi/bin/test`
- integration tests (against Minikube for now)
  - `export BOSH_KUBE_CPI_KUBE_CONFIG_PATH=~/.kubectl/config`
  - `ginkgo -r src/github.com/cppforlife/bosh-kubernetes-cpi/integration/`
- acceptance tests: `cd tests && ./run.sh` (against Minikube)

## TODO

### Must have

- file PR for director dns updates
- set pod priority
- determine draining plan of kube nodes
  - set pod disruption budget
  - eviction api: https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/#use-kubectl-drain-to-remove-a-node-from-service

### Nice to have

- disk migration (similar to docker cpi)
- lessen necessary perms on default container (what's agent doing?)
- works with gcr/ecr/ibm/harbor/on-prem registry
  - add authentication
- disable ntpdate updates
- gcr acceptance tests
- better error detection on vm creation before existing
  - non-pullable image?
- better error detection on disk creation
  - Warning   ProvisioningFailed  storageclass.storage.k8s.io "standard" not found (sl)
- automatically pick disk class default from a list?
- credential discovery for incluster vs outofcluster

### Enchancement

- setting AZs
  - through labels?
- automatically create anti affinity rules
- automatically create services?
- automatically create load balancer?
  - or update ports?
- manual networking via selected clusterIP services
- use daemon set to warm up stemcell loading?
  - when do we kick it off?
- do we need unique guid in front of heavy cid?
- bring back dead container if disk attach fails?
- minikube route to director?

# bosh-cpi-go

- error from cpiFactory.New (bosh-cpi-go)
- vmmeta stringmap
- vmenvgroup
- add set_disk_metadata
- add integration/testlib
