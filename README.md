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

## TODO

### Must have for production

- file PR for director dns updates
  - based on https://github.com/cloudfoundry/bosh/commit/98181d0a418382b8563ee74aced821932924b00a
- [in progress] determine draining plan of kube nodes
  - terminationGracePeriodSeconds
  - PDBs
    - fix StateApplier in cloudcheck_helper.rb
    - what determines container readiness/liveliness?
      - cat we use health.json?
    - how to optimize HM resurrection?
    - allow custom budgets (cross ig budget?)
    - which iaases use which drain command
      - drain timeout period? (eg `kubectl drain xxx --force --ignore-daemonsets --delete-local-data`)
    - associate controller owner with pods
    - director resurrection?
    - integration tests
    - how does hotswap affect budget?
    - how to remove unnecessary pdbs?

### Nice to have

- set pod priority
  - could be exposed via a cloud_property setting (via vm_extension)
- [dk] precompiled release
- non-root permissions on bosh container (what's agent doing?)
- use ibm registry
- use harbor registry
- use gcr registry in read-only for pulling
- gcr acceptance tests
- better error detection on vm creation before existing
  - non-pullable image?
- better error detection on disk creation
  - `Warning   ProvisioningFailed  storageclass.storage.k8s.io "standard" not found (sl)`
- automatically pick disk class default from a list?
  - configure as a global default cpi configuration
- automatically create registry secret with readonly pulling?
  - kube should only have read only access since it's not doing any pushing
- automatically create namespace?
- tag all resources with director/bosh label?
- [director] parallel disk creation

### Enhancement

- use daemon set to warm up stemcell loading?
  - during create_stemcell CPI call
  - delete during delete_stemcell CPI call
- do we need unique guid in front of heavy stemcell cid?
- bring back dead container if disk attach fails?
- minikube route to director?
- checked labels?
  - if cloud properties point to particular label, expect that somethign is bound to that label
  - fail if label does not point to anything?
  - which reosurces to look at (eg load balaancer vs ...)?
- update service's selector?
  - instead of labels, potentially ask for a service name and update its selector
  - effectively auto labeling
- credential discovery for incluster vs outofcluster
- bpmize pdbctrl (no acess to /var/run/...)

## Later

- disk migration (similar to docker cpi)
  - requires director changes
  - supporting attacgment of multiple disks to one vm
- disable ntpdate updates
  - currentl a hack in the cpi; ideally part of stemcell

# bosh-cpi-go

- error from cpiFactory.New (bosh-cpi-go)
- vmmeta stringmap
- vmenvgroup
- add set_disk_metadata
- add integration/testlib

# director go client

- delayed factory
