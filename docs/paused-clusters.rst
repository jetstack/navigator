Pausing Clusters
----------------

A cluster can be paused by setting ``spec.paused`` to ``true``.
When this is set, Navigator will not perform any actions on the cluster, allowing for manual intervention.

Paused clusters will have the following condition set:

.. code-block::

    Conditions:
      Last Transition Time:  2018-05-09T13:44:58Z
      Message:               Cluster is paused
      Reason:                ClusterPaused
      Status:                False
      Type:                  Progressing

When ``spec.paused`` is removed or set to ``false``, the corresponding condition will be set:

.. code-block::

    Conditions:
      Last Transition Time:  2018-05-09T13:46:23Z
      Message:               Cluster is resumed
      Reason:                ClusterResumed
      Status:                True
      Type:                  Progressing

