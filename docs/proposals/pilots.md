# Pilots

Pilots are application controllers that co-exist alongside each node in a
Navigator managed application. A pilot implements the logic required to
actually operate and interact with an application (as opposed to this logic
residing within Navigator itself).

There will be a Pilot for each application type, e.g. `pilot-elasticsearch`.
Navigator itself need not understand how a Pilot operates, so long as the Pilot
updates it's status through a corresponding `Pilot` resource (see the [pilots
resource proposal](pilot-resource.md) for more information).

## Goals

* **Clearly separate application and deployment logic** - at no point does
Navigator itself interact with the application process. Instead, the pilot
encodes the operational knowledge for interacting with the application. This
provides a clear separation of responsibilities.

## Non-goals

* **Provide a mechanism for deploying applications** - it is not expected that
a Pilot should be able to operate without the Navigator environment. It is
closely coupled to the navigator-apiserver, and relies upon the apiservers
versioning mechanisms for backwards compatibility.

## Responsibilities

* **Obtaining application & pilot configuration** - the pilot should obtain
configuration for the application from the corresponding Pilot resource stored
in the navigator-apiserver.

* **Launch and manage an application process** - the pilot is responsible for
spawning an application process after the configured preStart hooks have
completed.

* **Reporting application & pilot status to navigator-apiserver** - the status
of the pilot itself, as well as the application it manages should be
periodically reported to the `navigator-apiserver` in the `pilot.status` field.
This information is used by `navigator-controller` during cluster scale or
degraded health scenarios.

* **Provide complex health checks** - pilots should provide health check
endpoints that can check the health of the pilot, as well as the health of the
application instance. This helps bridge the gap between modern orchestration
systems requirements of health endpoints and the underlying application health
check facilities.

* **(Optional) Leader election** - if an application requires some or all
changes to cluster-wide state to be coordinator, they can use the navigator
API to perform leader election between Pilot resources. This is implemented in
a similar fashion to the leader election found in navigator-controller.

## Expected behaviour

The general operation of a pilot can be split into 5 phases:

### initializing

Initially, all instances of pilot are in this phase. In this phase, the pilot
is waiting for the initial Pilot resource for the pilot to exist in the
apiserver. Once the config has become available and is validated successfully,
it pilot will transition into the next phase.

### preStart

The preStart phase is before the application process has started, but after the
initial configuration has been validated. Hooks here will be executed according
to the rules they specify.

### postStart

### preStop

### postStop

## Building Pilots

In order to make building Pilots easier, we should create a generic
implementation the standardises common functionality between Pilots.

This new library will provide a number of services:

* **Machinery to set up a Sync based control loop** - all pilots are required
to watch for changes to their own Pilot resources, so the Pilot library should
make this easy. The will be configured through a developer-provided Sync
function that is specific to each Pilot. In future, adding a generic
implementation for this (similar to CRI) could be explored, that attempts to
more tightly define the interface for a Pilot reacting to changes to its own
resource).

* **Common utilities/functions to aid consistency between Pilots** - a core aim
of Navigator is to provide a reasonable level of consistency between the
behaviour of pilots. Functionality that is common between implementations
should be generalised and vendored into the genericpilot package for reuse.

* **Leader election** - if a Pilot requires some part of it to perform in a
leader elected fashion, the genericpilot package should provide a means for
an implementation to do this in a consistent way. By doing this, we can easily
present this information to an end-user in a consistent fashion for all Pilots.
It should be noted that this leader election may or may not be related to the
Pilots underlying application's leader election. Whether this is a reasonable
idea should be a decision left to the developer implementing the Pilot, taking
into account the Navigator API server's availability guarantees.
