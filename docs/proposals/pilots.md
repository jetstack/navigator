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

## Expected behaviour

The general operation of a pilot can be split into 4 phases:

* **preStart**

* **postStart**

* **preStop**

* **postStop**
