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

package escluster

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator"
)

func NewStrategy(typer runtime.ObjectTyper) esClusterStrategy {
	return esClusterStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*navigator.ElasticsearchCluster)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a ElasticsearchCluster.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), ESClusterToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchESCluster is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchESCluster(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// ESClusterToSelectableFields returns a field set that represents the object.
func ESClusterToSelectableFields(obj *navigator.ElasticsearchCluster) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type esClusterStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (esClusterStrategy) NamespaceScoped() bool {
	return true
}

func (esClusterStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
}

func (esClusterStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
}

func (esClusterStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (esClusterStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (esClusterStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (esClusterStrategy) Canonicalize(obj runtime.Object) {
}

func (esClusterStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
