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

All the C* nodes (pods) in a ``nodepool`` have the same configuration and the following sections describe the configuration options that are available.

.. note::
   Other than the following whitelisted fields, updates to nodepool configuration are not allowed:

   - ``replicas``
   - ``persistence``

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


Connecting to Cassandra
-----------------------

If you apply the YAML manifest from the example above,
Navigator will create a Cassandra cluster with three C* nodes running in three pods.
The IP addresses assigned to each C* node may change when pods are rescheduled or restarted, but there are stable DNS names which allow you to connect to the cluster.

Services and DNS Names
~~~~~~~~~~~~~~~~~~~~~~

Navigator creates two `headless services <https://kubernetes.io/docs/concepts/services-networking/service/#headless-services>`_ for every Cassandra cluster that it creates.
Each service has a corresponding DNS domain name:

#. The *nodes* service (e.g. ``cass-demo-nodes``) has a DNS domain name which resolves to the IP addresses of **all** the C* nodes in cluster (nodes 0, 1, and 2 in this example).
#. The *seeds* service (e.g. ``cass-demo-seeds``) has a DNS domain name which resolves to the IP addresses of **only** the `seed nodes <http://cassandra.apache.org/doc/latest/faq/index.html#what-are-seeds>`_ (node 0 in this example).

These DNS names have multiple HOST (`A`) records, one for each **healthy** C* node IP address.

.. note::
   The DNS server only includes `healthy <https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/>`_ nodes when answering requests for these two services.

The DNS names can be resolved from any pod in the Kubernetes cluster:

* If the pod is in the same namespace as the Cassandra cluster you need only use the left most label of the DNS name. E.g. ``cass-demo-nodes``.
* If the pod is in a different namespace you must use the fully qualified DNS name. E.g. ``cass-demo-nodes.my-namespace.svc.cluster.local``.

.. note::
   Read `DNS for Services and Pods <https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/>`_ for more information about DNS in Kubernetes.

TCP Ports
~~~~~~~~~

The C* nodes all listen on the following TCP ports:

#. **9042**: For CQL client connections.
#. **8080**: For Prometheus client connections.

Connect using a CQL Client
~~~~~~~~~~~~~~~~~~~~~~~~~~

Navigator configures all the nodes in a Cassandra cluster to listen on TCP port 9042 for `CQL client connections <http://cassandra.apache.org/doc/latest/cql/>`_.
And there are `CQL drivers for most popular programming languages <http://cassandra.apache.org/doc/latest/getting_started/drivers.html>`_.
Most drivers have the ability to connect to a single node and then discover all the other cluster nodes.

For example, you could use the `Datastax Python driver <http://datastax.github.io/python-driver/>`_ to connect to the Cassandra cluster as follows:

.. code-block:: python

   from cassandra.cluster import Cluster

   cluster = Cluster(['cass-demo-nodes'], port=9042)
   session = cluster.connect()
   rows = session.execute('SELECT ... FROM ...')
   for row in rows:
       print row

.. note::
   The IP address to which the driver makes the initial connection
   depends on the DNS server and operating system configuration.
