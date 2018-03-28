==================
Updating API types
==================

When updating API types, it's important to follow a protocol to ensure changes
are communicated clearly to users and all components support the new types
properly.

All PRs changing API types **require** a release note
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

PRs that touch API types should have a release note that clearly describes the
change, e.g:

.. code-block:: none

   ```release-note
   Added 'spec.minimumMasters' field to ElasticsearchCluster
   ```

All non-trivial fields should have a comment describing new fields/types
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Following godoc conventions allows us to generate API reference documentation
and publish information via swagger. All types and fields should have some form
of description in the form of a comment.

Run hack/update-client-gen.sh to regenerate clients etc
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

This step should also be enforced by CI testing. This will ensure deepcopies,
conversions, clientsets and informers are up to date. There is a script in
`hack/` that can be used:

.. code-block:: shell

    $ ./hack/update-client-gen.sh
    Generating deepcopy funcs
    Generating defaulters
    Generating clientset for navigator:v1alpha1 at github.com/jetstack/navigator/pkg/client/clientset
    Generating listers for navigator:v1alpha1 at github.com/jetstack/navigator/pkg/client/listers
    Generating informers for navigator:v1alpha1 at github.com/jetstack/navigator/pkg/client/informers
    Generating conversions

New functionality should be implemented with *fields* not annotations
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

If adding new functionality that is considered experimental or alpha to a beta
or stable API, add a new field to the resource and explicitly call out the
status of the field in its godoc:

.. code-block:: go
   :linenos:

   type FooCluster struct {
       ...
       // Replicas is the number of Foo cluster nodes to be created
       Replicas int64 `json:"replicas"`

       // EXPERIMENTAL: use of this field is considered experimental
       // TLS specifies the TLS configuration to use for the nodes in the Foo
       // cluster. If not specified, the cluster will be configured for insecure
       // connections.
       TLS *FooTLSConfig `json:"tls"`
   }

This provides type-safety and a mechanism for validation, as well as versioning
without extra complexity when promoting an old annotation to a field.
