# Pilot resource type

Navigator calls the manager of each instance in an application a Pilot. As
described in the [Pilot doc page](pilots.md), a Pilot is the
'glue' between Navigator and the applications it manages.

In order to store information about the state of each node in a Navigator
managed system (or 'cluster'), we introduce the Pilot resource type as part of
the `navigator` API group.

The Pilot resource will be namespaced, and will usually co-exist with a Pod
of the same name as itself.

## Goals

* Provide a mechanism to store metadata and state information about Pilot
instances, e.g.
    * The 'ready' status of a Pilot, through heartbeating a status condition

* Communicate pilot lifecycle events - for example an `inService` field could
be set to `false` in order to trigger the appropriate `Pilot` to migrate data
off of that particular instance. Similarly, the state of whether this process
has finished would then also be stored in the `status` block of the Pilot.

* Provide versioned configuration for each Pilot

## Non-goals

* Provide a synchronous API for interacting with Pilots

## Example manifest

```yaml
apiVersion: navigator.jetstack.io/v1alpha1
kind: Pilot
metadata:
  name: es-demo-data-0
  namespace: logging
spec:
  decommission: true
  elasticsearch:
    nodePool: mixed
    plugins:
    - name: io.fabric8:elasticsearch-cloud-kubernetes:5.2.2
status:
  elasticsearch:
    plugins:
    - name: io.fabric8:elasticsearch-cloud-kubernetes:5.2.2
      status: Unknown
  conditions:
  - type: Ready
    status: True
    lastTransitionTime: 2017-10-01T19:00:00+00
    reason: "ElasticsearchStarted"
    message: "Elasticsearch process has started"
  - type: Decommissioned
    status: False
    lastTransitionTime: 2017-10-01T19:00:00+00
    reason: "NodeDecommissioning"
    message: ""
```

## Scope of work

Almost all parts of Navigator will be touched by this feature. A brief
description of how this will affect these components follows.

### apiserver

The navigator-apiserver will need to add these new types to it's registries.
This should be fairly straightforward, and it can be added similar to how the
current `ElasticsearchCluster` resource type is added.

The `Pilot` resource type requires no special handling or validation for the
initial implementation.

### controller

#### Application controller changes

Each controller in `navigator-controller` will need to be modified to support
the new `Pilot` resource type in order to utilise it. Right now, we only have
support for `ElasticsearchCluster` resources, so this should be relatively
painless.

* **Creating `Pilot` resources** - an application controller should create `Pilot`
resources for each Pilot instance in applications it is managing. It is not
that these exist before the pilot is started, as the pilot should wait for its
corresponding resource to exist before starting.

* **Setting the `inService` field on a Pilot** - when a scale down of an
application or cluster occurs, a field should be set on the Pilots spec so
appropriate action can be taken by the pilot instances to decommission the node.
The exact specification for this feature is not set-in-stone, and the use of a
boolean `inService` field is used initially for it's simplicity.

#### Introduction of Pilot controller

A new Pilot control loop will be introduced in order to:

##### Pilot garbage collection

The pilot garbage collector should run periodically for each Pilot resource to
see the last time it was started, and if there is a corresponding pod with the
same name. If there is not for a configured duration, it is assumed the Pilot
has been removed due to a scale down event and the Pilot should be deleted.

If this is not the case, this should trigger the application controller to
create a new Pilot resource with the configuration required at the time of pod
creation.

##### Pilot status reconciliation

If a Pilot has not update it's Ready or Status field within a configured
duration, the status of the resource should be updated to set the fields to
a not ready state.

### pilots

Each pilot implementation will also need modifying quite extensively to support
the new type. The pilot is where the majority of the work will be required for
this feature.

#### Startup

On startup, pilots should wait until a Pilot resource with it's own name
exists before starting, as the Pilot resource can provide additional
configuration for the application.

#### Resource control loop

A pilot is similar to a kubelet, in that each one is coupled to a Pilot (or
Node in the kubelet's case) and must act in order to converge the state
specified in the resources `spec` field with the actual state of the node, and
report its status in the `status` field.

Pilots should watch for changes to themselves, and perform appropriate action
on their underlying applications in order to sync the 'actual' state of the
application & pilot with the desired state specified on the Pilot resource.
As some of these actions may actual need performing cluster wide (e.g.
excluding shard allocation on a node in elasticsearch requires a cluster-wide
config option to be updated), pilots can optionally elect a leader through
whatever means in order to converge the 'desired' and 'actual' state, however
it is the responsibility of the pilot instance for a corresponding Pilot
resource to update it's own status.
