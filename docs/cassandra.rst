Cassandra
=========

Example cluster definition
--------------------------

Example ``CassandraCluster`` resource:

.. include:: quick-start/cassandra-cluster.yaml
   :literal:

Node Pools
----------

The C* nodes in a Navigator ``cassandracluster`` are configured and grouped by rack and data center
and in Navigator, these groups of nodes are called ``nodepools``.

All the C* nodes (pods) in a ``nodepool`` have the same configuration and the following sections describe the configuration options that are available:

.. include:: configure-scheduler.rst

Cassandra Across Multiple Availability Zones
--------------------------------------------

With rack awareness
~~~~~~~~~~~~~~~~~~~

Navigator supports running Cassandra with
`rack and datacenter-aware replication <https://docs.datastax.com/en/cassandra/latest/cassandra/architecture/archDataDistributeReplication.html>`_
To deploy this, you must run a ``nodePool`` in each availability zone, and mark each as a separate Cassandra rack.

The
`nodeSelector <(https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector>`_
field of a nodePool allows scheduling the nodePool to a set of nodes matching labels.
This should be used with a node label such as
`failure-domain.beta.kubernetes.io/zone <https://kubernetes.io/docs/reference/labels-annotations-taints/#failure-domainbetakubernetesiozone>`_.

The ``datacenter`` and ``rack`` fields mark all Cassandra nodes in a nodepool as being located in that datacenter and rack.
This information can then be used with the
`NetworkTopologyStrategy <http://cassandra.apache.org/doc/latest/architecture/dynamo.html#network-topology-strategy>`_
keyspace replica placement strategy.
If these are not specified, Navigator will select an appropriate name for each: ``datacenter`` defaults to a static value, and ``rack`` defaults to the nodePool's name.

As an example, the nodePool section of a CassandraCluster spec for deploying into GKE in europe-west1 with rack awareness enabled:

.. code-block:: yaml

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

Without rack awareness
~~~~~~~~~~~~~~~~~~~~~~

Since the default rack name is equal to the nodepool name,
simply set the rack name to the same static value in each nodepool to disable rack awareness.

A simplified example:

.. code-block:: yaml

    nodePools:
    - name: "np-europe-west1-b"
      replicas: 3
      datacenter: "europe-west1"
      rack: "default-rack"
      nodeSelector:
        failure-domain.beta.kubernetes.io/zone: "europe-west1-b"
    - name: "np-europe-west1-c"
      replicas: 3
      datacenter: "europe-west1"
      rack: "default-rack"
      nodeSelector:
        failure-domain.beta.kubernetes.io/zone: "europe-west1-c"
    - name: "np-europe-west1-d"
      replicas: 3
      datacenter: "europe-west1"
      rack: "default-rack"
      nodeSelector:
        failure-domain.beta.kubernetes.io/zone: "europe-west1-d"

Managing Compute Resources for Cassandra Clusters
-------------------------------------------------

Each ``nodepool`` has a ``resources`` attribute which defines the resource requirements and limits for each C* node (pod) in the pool.

In the example above, each C* node (pod) in the ``nodepool`` named ``ringnodes`` will request half a CPU core and 2GiB of memory.

The ``resources`` field follows exactly the same specification as the Kubernetes Pod API
(``pod.spec.containers[].resources``).

See `Managing Compute Resources for Containers <https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/>`_ for more information.
