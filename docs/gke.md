## Stemcell

CPI is configured to use GCR with heavy warden stemcells.

## Director

See `deployments/gke/`

Nothing special.

## CF

See `deployments/gke-cf/`

Requirements:

- must select GKE Ubuntu image instead of COS
- must reboot nodes with `cgroup_enable=memory swapaccount=1` in `/boot/grub/grub.cfg`
  - we were not able to use update-grub utility
  - verify that `cat /proc/cmdline` contains above configuration
