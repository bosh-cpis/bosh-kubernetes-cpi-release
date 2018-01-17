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
- `src/src2` is docker registry libraries
- `src/src3` is copy of bosh-cron-release/src
