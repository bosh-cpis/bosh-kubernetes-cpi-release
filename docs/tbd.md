# Blue-green cluster upgrade

Challenges:

- how do you connect overlay networking into a new cluster
  - kubernetes may start giving out same cluster IPs -> configure different ranges
  - all depends on underlying overlay networking technology -> too different
