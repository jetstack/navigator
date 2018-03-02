# Navigator - self managed DBaaS on Kubernetes [![Build Status Widget]][Build Status]

Navigator is a Kubernetes extension for managing common stateful services on
Kubernetes. It is implemented as a custom apiserver that operates behind
[kube-aggregator](https://github.com/kubernetes/kube-aggregator) and introduces
a variety of new Kubernetes resource types.

As a result of this design, managing your services feels as natural as any
other resource in Kubernetes core. This means you can manage fine-grained
permissions via conventional RBAC rules, allowing you to offer popular but
complex services "as a Service" within your organisation.

For more in-depth information and to get started, jump to the [docs](https://navigator-dbaas.readthedocs.io).

Here's a quick demo of creating, scaling and deleting a Cassandra database:

![](sphinx-docs/images/demo.gif)

## Supported databases

Whilst we aim to support as many common applications as possible, it does take
a certain level of operational knowledge of the applications in question in
order to develop a pilot. Therefore, we'd like to reach out to others that are
interested in our efforts & would like to see a new application added (or
existing one improved!).

Please search for or create an issue for the application in question you'd like
to see a part of Navigator, and we can begin discussion on implementation &
planning.

| Name          | Version   | Status      | Notes                                                                             |
| ------------- | --------- | ----------- | --------------------------------------------------------------------------------- |
| Elasticsearch | 5.x       | Alpha       | [more info](https://navigator-dbaas.readthedocs.io/en/latest/elasticsearch.html)  |
| Cassandra     | 3.x       | Alpha       | [more info](https://navigator-dbaas.readthedocs.io/en/latest/cassandra.html)      |
| Couchbase     |           | Coming soon |                                                                                   |

## Credits

An open-source project by [Jetstack.io](https://www.jetstack.io/).

[Build Status Widget]: https://travis-ci.org/jetstack/navigator.svg?branch=master
[Build Status]: https://travis-ci.org/jetstack/navigator
