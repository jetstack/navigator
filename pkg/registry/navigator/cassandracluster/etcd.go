/*
Copyright 2017 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cassandracluster

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/registry"
)

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*registry.REST, *registry.REST, error) {
	strategy := NewStrategy(scheme)

	store := genericregistry.Store{
		NewFunc:       func() runtime.Object { return &navigator.CassandraCluster{} },
		NewListFunc:   func() runtime.Object { return &navigator.CassandraClusterList{} },
		PredicateFunc: MatchCassandraCluster,

		DefaultQualifiedResource: navigator.Resource("cassandraclusters"),

		CreateStrategy:          strategy,
		UpdateStrategy:          strategy,
		DeleteStrategy:          strategy,
		EnableGarbageCollection: true,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, nil, err
	}
	statusStore := store
	statusStore.UpdateStrategy = cassandraClusterStatusStrategy{strategy}

	return &registry.REST{
		Store:              &store,
		ResourceShortNames: []string{},
	}, &registry.REST{Store: &statusStore}, nil
}
