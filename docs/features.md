## Features

### AZs

You can set AZ via `zone` (and `region`) VM cloud properties. Appropriate node affinity will be set (see vm/affinity.go).

Possible enhancements: node affinity via labeling?

### Anti affinity rules

Anti affinity rules are automatically created for each instance group based on IG's `bosh.env.group` key (see vm/affinity.go)

### Manual networking

Manual networking can be used as long as ranges of cluster IPs are reserved for bosh use.

### Pod disruption budgets

CPI comes with a PDB controller that automatically creates PDBs for each instance group. By default only 1 allowed disruption is allowed per instance group. In cooperation with BOSH HM, Director will bring up evicted pods as they get terminated by Kubernetes.

Following PDB template is used:

```yaml
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: ???
spec:
  minAvailable: ??? # must be integer
  selector:
    matchLabels:
      bosh.io/group: ???
```

An example of a drain command:

```
$ kubectl drain <node-id> --force --ignore-daemonsets --delete-local-data
```

### TBDs

- automatically create services
- automatically create load balancer?
  - or update ports?
