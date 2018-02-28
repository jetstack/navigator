Cassandra Across Multiple Availability Zones
============================================

Navigator supports running Cassandra with
[rack and datacenter-aware replication](https://docs.datastax.com/en/cassandra/latest/cassandra/architecture/archDataDistributeReplication.html).
To deploy this, you must run a `nodePool` in each availability zone, and mark each as a separate Cassandra rack.

The
[`nodeSelector`](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)
field of a nodePool allows scheduling the nodePool to a set of nodes matching labels.
This should be used with a node label such as
[`failure-domain.beta.kubernetes.io/zone`](https://kubernetes.io/docs/reference/labels-annotations-taints/#failure-domainbetakubernetesiozone).

The `datacenter` and `rack` fields mark all Cassandra nodes in a nodepool as being located in that datacenter and rack.
This information can then be used with the
[`NetworkTopologyStrategy`](http://cassandra.apache.org/doc/latest/architecture/dynamo.html#network-topology-strategy)
keyspace replica placement strategy.
If these are not specified, Navigator will select an appropriate name for each: `datacenter` defaults to a static value, and `rack` defaults to the nodePool's name.

As an example, the nodePool section of a CassandraCluster spec for deploying into GKE in europe-west1 with rack awareness enabled:

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
