## Stemcell

CPI is configured to use GCR with heavy warden stemcells.

## Director

See `deployments/aks/`

Nothing special.

- LB creation failed initially

## CF

See `deployments/aks-cf/`

Requirements:

- must reboot nodes with `cgroup_enable=memory swapaccount=1` in `/boot/grub/grub.cfg`
  - verify that `cat /proc/cmdline` contains above configuration

Notes:

- AKS seems to be running 16.04.3 with 4.11 kernel
