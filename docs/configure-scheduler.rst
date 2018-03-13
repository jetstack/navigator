Configure scheduler type
------------------------

If a custom scheduler type is required (for example if you are deploying with Portworx or another storage provider), this can be set on either each nodepool:

.. code-block:: yaml

    spec:
      schedulerName: "fancy-scheduler"
      nodePools:
      - name: "ringnodes-1"
      - name: "ringnodes-2"

or for an entire cluster:

.. code-block:: yaml

    spec:
      nodePools:
      - name: "ringnodes-1"
        schedulerName: "fancy-scheduler"
      - name: "ringnodes-2"
        schedulerName: "fancy-scheduler"

The nodepool field takes precedent, falling back to the cluster value if it is defined, then the default scheduler.
