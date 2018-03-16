Elasticsearch
=============

Example cluster definition
--------------------------

Example ``ElasticsearchCluster`` resource:

.. include:: quick-start/es-cluster-demo.yaml
   :literal:

Node Pools
----------

The Elasticsearch nodes in a Navigator ``ElasticsearchCluster`` are configured and grouped by role
and in Navigator, these groups of nodes are called ``nodepools``.

.. note::
   Other than the following whitelisted fields, updates to nodepool configuration are not allowed:

   - ``replicas``
   - ``persistence``

.. include:: configure-scheduler.rst

.. include:: managing-compute-resources.rst

.. _system-configuration-elastic-search:

System Configuration for Elasticsearch Nodes
--------------------------------------------

Elasticsearch requires `important system configuration settings <https://www.elastic.co/guide/en/elasticsearch/reference/current/system-config.html>`_ to be applied globally on the host operating system.

You must either ensure that Navigator is running in a Kubernetes cluster where all the nodes have been configured this way.
Or you could use `node labels and node selectors <https://kubernetes.io/docs/concepts/configuration/assign-pod-node/>`_ to ensure that the pods of an Elasticsearch cluster are only scheduled to nodes with the required configuration.

See `Using Sysctls in a Kubernetes Cluster <https://kubernetes.io/docs/concepts/cluster-administration/sysctl-cluster/>`_,
and `Taints and Tolerations <https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/>`_ for more information.

One way to apply these settings is to deploy a ``DaemonSet`` that runs the configuration commands from within a privileged container on each Kubernetes node.
Here's a simple example of such a ``DaemonSet``:

.. code-block:: bash

   $ kubectl apply -f docs/quick-start/sysctl-daemonset.yaml

.. include:: quick-start/sysctl-daemonset.yaml
   :literal:

:download:`docs/quick-start/sysctl-daemonset.yaml <quick-start/sysctl-daemonset.yaml>`
