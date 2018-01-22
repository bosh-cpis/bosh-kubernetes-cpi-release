## Stemcell

CPI is configured to use minikube's Docker daemon with heavy warden stemcells.

## Director

Steps:

- `minikube start`
- `cd ./deployments/minikube/`
- `kubectl create -f ../generic/ns.yml`
- `./run`

Requirements:

- have to use NodePorts for now

## CF

Requirements:

- had to load btrfs kernel module on the minikube vbox vm
