# BOSH Kubernetes CPI

The BOSH Kubernetes CPI allows BOSH to manage deploy BOSH workloads such as CF onto Kubernetes clusters.

## Use with Kube environments

- [GKE](docs/gke.md)
- [IBM](docs/ibm.md)
- [AKS (Azure)](docs/aks.md)
- [Minikube](docs/minikube.md)

## Development

- unit tests
  - `./src/github.com/cppforlife/bosh-kubernetes-cpi/bin/test`
- integration tests (against Minikube for now)
  - `export BOSH_KUBE_CPI_KUBE_CONFIG_PATH=~/.kube/config`
  - `ginkgo -r src/github.com/cppforlife/bosh-kubernetes-cpi/integration/`
- acceptance tests: `cd tests && ./run.sh` (against Minikube)

## TODO

### Must have

- file PR for director dns updates
  - based on https://github.com/cloudfoundry/bosh/commit/98181d0a418382b8563ee74aced821932924b00a
- set pod priority
- determine draining plan of kube nodes
  - set pod disruption budget
  - eviction api: https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/#use-kubectl-drain-to-remove-a-node-from-service
  - terminationGracePeriodSeconds

### Nice to have

- precompiled release
- disk migration (similar to docker cpi)
  - requires director changes
- lessen necessary perms on default container (what's agent doing?)
- works with gcr/ecr/ibm/harbor/on-prem registry
  - add authentication
- disable ntpdate updates
- gcr acceptance tests
- better error detection on vm creation before existing
  - non-pullable image?
- better error detection on disk creation
  - `Warning   ProvisioningFailed  storageclass.storage.k8s.io "standard" not found (sl)`
- automatically pick disk class default from a list?
- automatically create registry secret with readonly pulling?
- automatically create namespace?

### Enhancement

- use daemon set to warm up stemcell loading?
  - when do we kick it off?
- do we need unique guid in front of heavy cid?
- bring back dead container if disk attach fails?
- minikube route to director?
- checked labels?
- update service's selector?
- credential discovery for incluster vs outofcluster

# bosh-cpi-go

- error from cpiFactory.New (bosh-cpi-go)
- vmmeta stringmap
- vmenvgroup
- add set_disk_metadata
- add integration/testlib
