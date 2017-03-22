# Colonel - easy DBaaS on Kubernetes

Colonel is a centralised controller for managing common stateful services on Kubernetes.
It provides a framework for building control loops that manage cluster resources based on
ThirdPartyResources defining the particular service you want to run.

Currently, it only supports Elasticsearch 5.2.x, however support for more popular services
is coming _very_ soon!

## Architecture

> TODO: Diagram here

## Quick-start Elasticsearch

Here we're going to deploy a distributed and scalable Elasticsearch cluster using the examples
provided in this repository. This will involve first deploying Colonel, and then creating
an `ElasticsearchCluster` resource. All management of the Elasticsearch cluster will be through
changes to the ElasticsearchCluster manifest.

1) Install Colonel by creating the deployment manifest:

```bash
$ kubectl create -f examples/deploy.yaml
```

2) Create a new ElasticsearchCluster:

```bash
$ kubectl create -f examples/es-cluster-example.yaml
```

This will deploy a multi-node Elasticsearch cluster, split into nodes of 3 roles: master, client (ingest) and data.
There will be 4 data nodes, each with a 10GB PV, 2 client nodes, and 3 master nodes.

3) Scale the data nodes:

Scaling the nodes can be done by modifying your ElasticsearchCluster manifest. Currently this is only
possible using `kubectl replace`, due to bugs with the way ThirdPartyResource's are handled in kubectl 1.5.

Edit your manifest & **increase** the number of replicas in the `data` node pool, then run:

```bash
$ kubectl replace -f examples/es-cluster-example.yaml
```

You should see new data nodes being added into your cluster gradually. Once all are in the Running state, we can try
a scale down. Do the same as before, but instead reduce the number of replicas in the `data` node pool. Then run a
`kubectl replace` again:

```bash
$ kubectl replace -f examples/es-cluster-example.yaml
```

Upon scale-down, the Elasticsearch nodes will mark themselves as non-allocatable. This will trigger Elasticsearch to
re-allocate any shards currently on the nodes being scaled down, meaning your data will be safely relocated within the
cluster. It should be noted that this may not be possible depending on the replication policies set on your indices, and in
this case, the node will hang in the 'Terminating' state until eventually being force-killed (sent a SIGKILL signal).

It is up to the operator to ensure that indices policies are achievable given the ElasticsearchCluster topology.