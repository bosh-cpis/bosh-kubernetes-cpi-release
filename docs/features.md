## Features

### AZs

You can set AZ via `zone` (and `region`) VM cloud properties. Appropriate node affinity will be set (see vm/affinity.go).

Possible enhancements: node affinity via labeling?

### Anti affinity rules

Anti affinity rules are automatically created for each instance group based on IG's `bosh.env.group` key (see vm/affinity.go)

### Manual networking

Manual networking can be used as long as ranges of cluster IPs are reserved for bosh use.

### TBDs

- automatically create services
- automatically create load balancer?
  - or update ports?
