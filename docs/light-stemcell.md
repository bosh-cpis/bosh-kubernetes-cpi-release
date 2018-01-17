# Light Stemcell

Can import heavy regular Warden stemcell directly into Minikube via docker:

```
$ eval $(minikube docker-env)
$ cat ~/Downloads/bosh-stemcell-3468.15-warden-boshlite-ubuntu-trusty-go_agent/image | docker import - bosh.io/stemcells:533
```

Use `stemcell/` to build 3468.533 test light stemcell and use that with BOSH.
