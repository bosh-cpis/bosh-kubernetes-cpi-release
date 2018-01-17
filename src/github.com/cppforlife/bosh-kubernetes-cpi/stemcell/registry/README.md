## References

- https://docs.docker.com/registry/spec/api/#base

## Example endpoints

- `curl -vvv http://192.168.99.100:5000/v2/`
- `curl -vvv http://192.168.99.100:5000/v2/_catalog`
- `curl -vvv http://192.168.99.100:5000/v2/_catalog/bosh.io/stemcells`
- `docker pull localhost:5000/bosh.io/stemcells@sha256:692b59d12ef7a136f51300bb1387913068c013ab58268d7d99ccd5b773075e86`
