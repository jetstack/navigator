# CouchbaseCluster

```yaml
apiVersion: navigator.jetstack.io/v1alpha1
kind: CouchbaseCluster
metadata:
  name: demo
spec:
  # this version number will be added as an annotation to each pod
  version: '4.5.1'
  # pilot image to use
  image:
    repository: jetstackexperimental/pilot-couchbase
    tag: latest
    pullPolicy: IfNotPresent

  # a list of nodepools that form the cluster
  nodePools:
  - name: data
    replicas: 1
    # nodes in a pool can be assigned roles, eg. 'data', 'query', 'index'
    roles:
    - data
```
