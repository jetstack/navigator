package v1

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

// We must register our custom API types with the apimachinery runtime.
// This file builds a new Scheme, and registers it.

var (
	// SchemeBuilder that will register marshal v1 types to a scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme of this APIs SchemeBuilder
	AddToScheme = SchemeBuilder.AddToScheme
)

// GroupName for this API
const GroupName = "alpha.marshal.io"

// Version for this API
const Version = "v1"

// SchemeGroupVersion for this API
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

// Kind returns a schema.GroupKind for the given kind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource returns a schema.GroupResource for the given resource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the v1 types to a scheme
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ElasticsearchCluster{},
		&ElasticsearchClusterList{},
		&v1.ListOptions{},
	)
	return nil
}

// Install registers the API group and adds types to a scheme
func Install(groupFactoryRegistry announced.APIGroupFactoryRegistry, registry *registered.APIRegistrationManager, scheme *runtime.Scheme) {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:                  SchemeGroupVersion.Group,
			VersionPreferenceOrder:     []string{SchemeGroupVersion.Version},
			ImportPrefix:               "github.com/jetstack-experimental/navigator/pkg/api/v1",
			AddInternalObjectsToScheme: AddToScheme,
		},
		announced.VersionToSchemeFunc{
			SchemeGroupVersion.Version: AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme); err != nil {
		panic(err)
	}
}

func init() {
	Install(api.GroupFactoryRegistry, api.Registry, api.Scheme)
}
