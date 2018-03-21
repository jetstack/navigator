Quick-start
===========

Here we're going to deploy a distributed and scalable Elasticsearch cluster using the examples provided in this repository.
This will involve first deploying Navigator, and then creating an ``ElasticsearchCluster`` resource.
All management of the Elasticsearch cluster will be through changes to the ElasticsearchCluster manifest.

1) Install Navigator using `Helm <https://github.com/kubernetes/helm>`_
-----------------------------------------------------------------------

.. code-block:: bash

    $ helm install contrib/charts/navigator --name navigator --namespace navigator --wait

You should see the Navigator service start in the ``navigator`` namespace:

.. code-block:: bash

    $ kubectl get pod -n navigator
    NAME                        READY     STATUS         RESTARTS   AGE
    navigator-745449320-dcgms   1/1       Running        0          30s

2) Prepare your Kubernetes nodes
--------------------------------

Elasticsearch requires certain `important system configuration settings <https://www.elastic.co/guide/en/elasticsearch/reference/current/system-config.html>`_ to be configured on the host operating system i.e. on the Kubernetes node.
For this demonstration, it should be sufficient to run ``sysctl -w vm.max_map_count=262144``, which `increases a particular virtual memory limit <https://www.elastic.co/guide/en/elasticsearch/reference/current/vm-max-map-count.html>`_.
You can quickly run the command on *all* your Kubernetes nodes by installing the following ``DaemonSet``:

.. code-block:: bash

  $ kubectl apply -f docs/quick-start/sysctl-daemonset.yaml

Or you can log into each node and run the command by hand.

See :ref:`system-configuration-elastic-search` for more information.

3) Create an Elasticsearch cluster
----------------------------------

.. code-block:: bash

    $ kubectl create -f docs/quick-start/es-cluster-demo.yaml

This will deploy a multi-node Elasticsearch cluster, split into nodes of 3 roles: master, client (ingest) and data.
There will be 4 data nodes, each with a 10GB PV, 2 client nodes, and 3 master nodes.
All of the options you may need for configuring your cluster are documented on the supported types page.

.. code-block:: bash

    $ kubectl get pod
    NAME                              READY     STATUS    RESTARTS   AGE
    es-demo-client-3995124321-5rc6g   1/1       Running   0          7m
    es-demo-client-3995124321-9zrv9   1/1       Running   0          7m
    es-demo-data-0                    1/1       Running   0          7m
    es-demo-data-1                    1/1       Running   0          5m
    es-demo-data-2                    1/1       Running   0          3m
    es-demo-data-3                    1/1       Running   0          1m
    es-demo-master-554549909-00162    1/1       Running   0          7m
    es-demo-master-554549909-pp557    1/1       Running   0          7m
    es-demo-master-554549909-vjgrt    1/1       Running   0          7m

4) Scale the data nodes
-----------------------

Scaling the nodes can be done by modifying your ElasticsearchCluster manifest.
Currently this is only possible using ``kubectl replace``, due to bugs with the way ThirdPartyResource's are handled in kubectl 1.5.

Edit your manifest and **increase** the number of replicas in the ``data`` node pool, then run:

.. code-block:: bash

    $ kubectl replace -f examples/es-cluster-example.yaml
    $ kubectl get pod
    NAME                              READY     STATUS    RESTARTS   AGE
    es-demo-client-3995124321-5rc6g   1/1       Running   0          9m
    es-demo-client-3995124321-9zrv9   1/1       Running   0          9m
    es-demo-data-0                    1/1       Running   0          9m
    es-demo-data-1                    1/1       Running   0          7m
    es-demo-data-2                    1/1       Running   0          5m
    es-demo-data-3                    1/1       Running   0          3m
    es-demo-data-4                    0/1       Running   0          29s
    es-demo-master-554549909-00162    1/1       Running   0          9m
    es-demo-master-554549909-pp557    1/1       Running   0          9m
    es-demo-master-554549909-vjgrt    1/1       Running   0          9m

You should see new data nodes being added into your cluster gradually.
Once all are in the Running state, we can try a scale down.
Do the same as before, but instead reduce the number of replicas in the ``data`` node pool.
Then run a ``kubectl replace`` again:

.. code-block:: bash

    $ kubectl replace -f examples/es-cluster-example.yaml
    $ kubectl get pod
    NAME                              READY     STATUS        RESTARTS   AGE
    es-demo-client-3995124321-5rc6g   1/1       Running       0          10m
    es-demo-client-3995124321-9zrv9   1/1       Running       0          10m
    es-demo-data-0                    1/1       Running       0          10m
    es-demo-data-1                    1/1       Running       0          8m
    es-demo-data-2                    1/1       Running       0          6m
    es-demo-data-3                    1/1       Running       0          4m
    es-demo-data-4                    1/1       Terminating   0          2m
    es-demo-master-554549909-00162    1/1       Running       0          10m
    es-demo-master-554549909-pp557    1/1       Running       0          10m
    es-demo-master-554549909-vjgrt    1/1       Running       0          10m

Upon scale-down, the Elasticsearch nodes will mark themselves as non-allocatable.
This will trigger Elasticsearch to re-allocate any shards currently on the nodes being scaled down, meaning your data will be safely relocated within the cluster.
