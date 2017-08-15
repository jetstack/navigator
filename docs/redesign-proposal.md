# Navigator design

Navigator as a system intends to provide a way to easily provision, monitor and
manage complex systems on top of cluster schedulers, initially targetting
Kubernetes and OpenShift.

## Architecture

All communication between components of the Navigator system should communicate
through the `navigator-apiserver`. This allows for a consistent audit trail,
recorded in a central place. Monitoring services such as `kube-state-metrics`
can then be used to track changes to the contents of the API server to be
consumed by something like Prometheus.

## Pilots

Databases run by Navigator are wrapped by helper processes called 'Pilots'.

Pilots contain the operational knowledge needed to operate their respective
databases with the assistance of the `navigator-apiserver`.

Each Pilot runs as PID 1 within the main application container of the
application it 'manages'. This allows us to closely and carefuly monitor the
lifecycle events and signals that the subprocess receives. This will allow us
to run Pilot in environments other than Kubernetes, despite the individual
properties of how these systems work.

Because the Pilot runs before the actual application has started, it is able to
perform any preflight or configuration checks at an appropriate time. This may
include waiting for other Pilots to start, or determining a leader (with the
help of the `navigator-apiserver`) that should begin the bootstrapping of a new
cluster for that application. This also applies at exit time. The Pilot is able
to trap signals sent to the container and ensure the underlying application
exits cleanly, and if it doesn't, report this diagnostic information back to
the `navigator-apiserver`.

Pilots should watch the `navigator-apiserver` for changes to resources it
requires, and use the apiserver to communicate with other instances of Pilot
within the same application cluster (eg. within the same Elasticsearch
cluster). This may include things like notifications that a scale down event
has been requested, or even used to elect a 'leader' between the Pilots in the
event one is required for the proper functioning of that application.

Each Pilot should also be able to expose a metrics endpoint to be scraped by
systems like Prometheus. This could be implemented via some of the existing
exporters for common applications, or directly within the Pilot itself if
required.

## navigator-controller-manager

The `navigator-controller-manager` should implement all of the control features
that are required for the proper functioning of the system. It should behave
very much like the `kube-controller-manager`. If, for example, a user creates
an `ElasticsearchCluster` resource, the `navigator-controller-manager` will
create whatever corresponding resources are required in the target orchestrator
in order to fulfill the request. It should then report status information of
its actions back to the `navigator-apiserver`. Additional status information
may also be provided by the Pilots themselves in this case, and in fact the
`controller-manager` may also consequently perform *other* actions as a result.

If a user were to modify the size of a node pool within an
ElasticsearchCluster, the controller-manager would consequently notify the
relevant Pilots of this event. It could then wait until a successful data
migration has occured before actually scaling down the appropriate
StatefulSet/Deployment. The precise semantics of *how* the controller-manager
should communicate this to the Pilots is undefined, but however it is done it
should be via the API server (either by updating the `status` field of the
ElasticsearchCluster resource, or by creating an additional `ScaleEvent`
resource to be fulfilled by the Pilots).

Only one instance of the controller-manager should run at a given time, and
they should leader-elect amongst themselves using either the
`navigator-apiserver` or some other leader election primitive provided by the
target orchestrator.

## navigator-apiserver

The apiserver component acts as the primary API surface and data storage for
all resources types that we define as part of our API. It's design should be
heavily modelled upon that of `kube-apiserver`, in that it should not perform
much actual control logic itself.

The `navigator-apiserver` will understand multiple versions of our types, thus
allowing us to support `v1` configuration for types even with the newest
versions of the `navigator-controller-manager` by losslessly converting between
the old and new version types. More information about this can be found in
my talk on API internals, as well as on the upcoming blog post.

Like the `kube-apiserver`, multiple instances of it may run at one time to
provide scalability on both reads and writes. The underlying data store may be
`etcd`, Kubernetes CRDs or OpenShift TPRs. It is undecided which will be
implemented first, however good support for `etcd` is provided in the
`k8s.io/apiserver` library.
