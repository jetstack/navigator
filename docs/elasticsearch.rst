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
