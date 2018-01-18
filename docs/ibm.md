## Stemcell

CPI is configured to use GCR with heavy warden stemcells (todo - add ibm registry).

## Director

See `deployments/ibm/`

Requirements:

- use standard cluster
  - we did not bother to make NodePort configuration work on free version
- had to use `ibm-file-gold` for acceptable disk IO
- had to get a subnet for public LoadBalancer

## CF

See `deployments/ibm-cf/`

Nothing special.
