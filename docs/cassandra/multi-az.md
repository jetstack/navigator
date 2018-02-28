Cassandra across multiple availability zones
============================================

Navigator supports running Cassandra with rack and datacenter-aware
replication. To deploy this, you must run a `nodePool` in each availability
zone, and mark each as a separate Cassandra rack.

The `nodeSelector` field of a nodePool allows scheduling the nodePool to a set
of nodes matching labels. This should be used with a node label such as
`failure-domain.beta.kubernetes.io/zone`.

The `datacenter` and `rack` fields mark all Cassandra nodes in a nodepool as
being located in that datacenter and rack. This information can then be used
with the `NetworkTopologyStrategy` keyspace replica placement strategy.

As an example, the nodePool section of a CassandraCluster spec for deploying
into GKE in europe-west1:

```yaml
  nodePools:
  - name: "np-europe-west1-b"
    replicas: 3
    datacenter: "europe-west1"
    rack: "europe-west1-b"
    nodeSelector:
      failure-domain.beta.kubernetes.io/zone: "europe-west1-b"
    persistence:
      enabled: true
      size: "5Gi"
      storageClass: "default"
  - name: "np-europe-west1-c"
    replicas: 3
    datacenter: "europe-west1"
    rack: "europe-west1-c"
    nodeSelector:
      failure-domain.beta.kubernetes.io/zone: "europe-west1-c"
    persistence:
      enabled: true
      size: "5Gi"
      storageClass: "default"
  - name: "np-europe-west1-d"
    replicas: 3
    datacenter: "europe-west1"
    rack: "europe-west1-d"
    nodeSelector:
      failure-domain.beta.kubernetes.io/zone: "europe-west1-d"
    persistence:
      enabled: true
      size: "5Gi"
      storageClass: "default"
```
