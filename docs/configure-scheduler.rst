Configure Scheduler Type
------------------------

If a `custom scheduler <https://kubernetes.io/docs/tasks/administer-cluster/configure-multiple-schedulers/>`_ type is required
(for example if you are deploying with `stork <https://docs.portworx.com/scheduler/kubernetes/stork.html>`_ or another storage provider),
this can be set on each nodepool:

.. code-block:: yaml

    spec:
      nodePools:
      - name: "ringnodes-1"
        schedulerName: "fancy-scheduler"
      - name: "ringnodes-2"
        schedulerName: "fancy-scheduler"

If the nodepool field is not specified, the default scheduler is used.
