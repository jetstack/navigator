package v1

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

const GroupName = "alpha.marshal.io"
const Version = "v1"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

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
			ImportPrefix:               "gitlab.jetstack.net/marshal/colonel/pkg/api/v1",
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
