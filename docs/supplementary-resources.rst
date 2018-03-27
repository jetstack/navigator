Pilots
------

Navigator creates one ``Pilot`` resource for every database node.
``Pilot`` resources have the same name and name space as the ``Pod`` for the corresponding database node.
The ``Pilot.Spec`` is read by the pilot process running inside a ``Pod`` and contains its desired configuration.
The ``Pilot.Status``  is updated by the pilot process and contains the discovered state of a single database node.

Other Supplementary Resources
-----------------------------

Navigator will also create a number of supplementary resources for each cluster.
For example it will create a ``serviceaccount``, a ``role`` and a ``rolebinding``
so that pilot pods in a cluster have read-only access the API resources containing cluster configuration,
and so that pilot pods can update the status of their corresponding ``Pilot`` resource and leader election ``configmap``.
