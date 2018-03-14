Managing Compute Resources for Clusters
---------------------------------------

Each ``nodepool`` has a ``resources`` attribute which defines the resource requirements and limits for each database node (pod) in that pool.

In the example above, each database node will request 0.5 CPU core and 2GiB of memory,
and will be limited to 1 CPU core and 3GiB of memory.

The ``resources`` field follows exactly the same specification as the Kubernetes Pod API
(``pod.spec.containers[].resources``).

See `Managing Compute Resources for Containers <https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/>`_ for more information.
