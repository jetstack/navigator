# Navigator - managed DBaaS on Kubernetes [![Build Status Widget]][Build Status]

Navigator is a Kubernetes extension for managing common stateful services on
Kubernetes. It is implemented as a custom apiserver that operates behind
[kube-aggregator](https://github.com/kubernetes/kube-aggregator) and introduces
a variety of new Kubernetes resource types.

As a result of this design, managing your services feels as natural as any
other resource in Kubernetes core. This means you can manage fine-grained
permissions via conventional RBAC rules, allowing you to offer popular but
complex services "as a Service" within your organisation.

To get started, jump to the [quick-start guide](docs/quick-start).

## Design

As well as following "the operator model", Navigator additionally introduces
the concept of 'Pilots' - small 'nanny' processes that run inside each pod
in your application deployment. These Pilots are responsible for managing the
lifecycle of your underlying application process (e.g. an Elasticsearch JVM
process) and periodically report state information about the individual node
back to the Navigator API.

By separating this logic into it's own binary that is run alongside each node,
in certain failure events the Pilot is able to intervene in order to help
prevent data loss, or otherwise update the Navigator API with details of the
failure so that navigator-controller can take action to restore service.

Navigator has a few unique traits that differ from similar projects (such as
elasticsearch-operator, etcd-operator etc).

- **navigator-apiserver** - this takes on a similar role to `kube-apiserver`.
It is responsible for storing and coordinating all of the state stored for
Navigator. It requires a connection to an etcd cluster in order to do this. In
order to make Navigator API types generally consumable to users of your cluster,
it registers itself with kube-aggregator. It performs validation of your
resources, as well as performing conversions between API versions which allow
us to maintain a stable API without hindering development.

- **navigator-controller** - the controller is akin to `kube-controller-manager`.
It is responsible for actually realising your deployments within the Kubernetes
cluster. It can be seen as the 'operator' for the various applications
supported by `navigator-apiserver`.

- **pilots** - the pilot is responsible for managing each database process.
  Currently Navigator has two types: `pilot-elasticsearch` and
  `pilot-cassandra`.

## Architecture

![alt text](docs/arch.jpg)

## Supported applications

Whilst we aim to support as many common applications as possible, it does take
a certain level of operational knowledge of the applications in question in
order to develop a pilot. Therefore, we'd like to reach out to others that are
interested in our efforts & would like to see a new application added (or
existing one improved!).

Please search for or create an issue for the application in question you'd like
to see a part of Navigator, and we can begin discussion on implementation &
planning.

| Name          | Version   | Status      | Notes                                                       |
| ------------- | --------- | ----------- | ----------------------------------------------------------- |
| Elasticsearch | 5.x       | Alpha       | [more info](docs/supported-types/elasticsearch-cluster.md)  |
| Cassandra     | 3.x       | Alpha       | [more info](docs/supported-types/cassandra-cluster.md)      |
| Couchbase     |           | Coming soon |                                                             |

## Links

* [Quick-start](docs/quick-start)
* [Developing quick-start](docs/developing.md)
* [Resource types](docs/supported-types/README.md)
  * [ElasticsearchCluster](docs/supported-types/elasticsearch-cluster.md)


## E2E Testing

Navigator has an end-to-end test suite which verifies that Navigator can be
installed [as documented in the quick start guide](docs/quick-start).  The
tests are run on a Minikube cluster.  Run the tests using the following
sequence of commands:

```
minikube start
# This ensures that the Docker image will be built in the Minikube VM
eval $(minikube docker-env)
# Override the Docker image tag so that it is built as :latest
# (the tag used in the documented deployment)
# XXX: This is a hack.
# Better if we had a helm chart in the documentation,
# so that we could provide an alternative navigator image and tag.
make BUILD_TAG=latest e2e-test
```

## Credits

An open-source project by [Jetstack.io](https://www.jetstack.io/).

[Build Status Widget]: https://travis-ci.org/jetstack/navigator.svg?branch=master
[Build Status]: https://travis-ci.org/jetstack/navigator
