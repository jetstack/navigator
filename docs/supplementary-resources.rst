Other Supplementary Resources
-----------------------------

Navigator will also create a number of supplementary resources for each cluster.
For example it will create a ``serviceaccount``, a ``role`` and a ``rolebinding``
so that pilot pods in a cluster have read-only access the API resources containing cluster configuration,
and so that pilot pods can update the status of their corresponding ``Pilot`` resource and leader election ``configmap``.
