# ElasticsearchCluster

```yaml
apiVersion: alpha.marshal.io/v1
kind: ElasticsearchCluster
metadata:
  name: demo
spec:
  # this version number will be added as an annotation to each pod
  version: '5.2.2'
  # a list of additional plugins to install
  plugins:
  - name: "io.fabric8:elasticsearch-cloud-kubernetes:5.2.2"

  # custom sysctl's to set. These are set by privileged init containers.
  sysctl:
  - vm.max_map_count=262144

  # lieutenant image to use
  image:
    repository: jetstackexperimental/lieutenant-elastic-search
    tag: master-907
    pullPolicy: Always
    ## This sets the group of the persistent volume created for
    ## the data nodes. This must be the same as the user that elasticsearch
    ## runs as within the container.
    fsGroup: 1000

  # a list of nodepools that form the cluster
  nodePools:
  - name: data
    replicas: 3

    # nodes in a pool can be assigned roles, eg. 'data', 'client' and 'master'
    roles:
    - data
    
    resources:
      requests:
        cpu: '500m'
        memory: 2Gi
      limits:
        cpu: '1'
        memory: 3Gi
    
    state:
      # stateful defines whether to create a StatefulSet or a Deployment
      stateful: true

      # persistent sets persistent storage config
      persistence:
        # enabled will only have an effect if stateful is also true. If true,
        # the persistentVolumeClaims block of the StatefulSet will be set
        enabled: true
        # size of the volume
        size: 10Gi
        # storageClass of the volume
        storageClass: "fast"

  - name: client
    replicas: 2

    roles:
    - client
    
    resources:
      requests:
        cpu: '1'
        memory: 2Gi
      limits:
        cpu: '2'
        memory: 4Gi

  - name: master
    replicas: 3

    roles:
    - master
    
    resources:
      requests:
        cpu: '1'
        memory: 2Gi
      limits:
        cpu: '2'
        memory: 4Gi
```