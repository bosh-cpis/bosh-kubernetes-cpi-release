## Stemcell

CPI is configured to use GCR with heavy warden stemcells.

## Director

Steps:

- create GKE cluster
- `cd ./deployments/gke/`
- build `./kubeconfig`
- `kubectl create -f ../generic/ns.yml`
- `kubectl create -f ../generic/lb.yml`
- update `./run` with API and LB IPs
- create GCR credential and save to `./gcr-password.json`
- `kubectl create secret -n bosh docker-registry regsecret -n bosh --docker-server=https://gcr.io --docker-username=_json_key --docker-password="$(cat gcr-password.json)" --docker-email foo` to create secret for pulling images
- `./run`

Nothing special.

## CF

See `deployments/gke-cf/`

Requirements:

- must select GKE Ubuntu image instead of COS
- must reboot nodes with `cgroup_enable=memory swapaccount=1` in `/boot/grub/grub.cfg`
  - we were not able to use update-grub utility
  - verify that `cat /proc/cmdline` contains above configuration
