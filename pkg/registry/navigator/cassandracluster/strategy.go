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
	"fmt"

	"github.com/golang/glog"
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

func NewStrategy(typer runtime.ObjectTyper) cassandraClusterStrategy {
	return cassandraClusterStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*navigator.CassandraCluster)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a CassandraCluster.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), CassandraClusterToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchCassandraCluster is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchCassandraCluster(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// CassandraClusterToSelectableFields returns a field set that represents the object.
func CassandraClusterToSelectableFields(obj *navigator.CassandraCluster) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type cassandraClusterStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (cassandraClusterStrategy) NamespaceScoped() bool {
	return true
}

func (cassandraClusterStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
}

func (cassandraClusterStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
}

func (cassandraClusterStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (cassandraClusterStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (cassandraClusterStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (cassandraClusterStrategy) Canonicalize(obj runtime.Object) {
}

func (cassandraClusterStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

// implements interface RESTUpdateStrategy. This implementation validates updates to
// instance.Status updates only and disallows any modifications to the instance.Spec.
type cassandraClusterStatusStrategy struct {
	cassandraClusterStrategy
}

func (cassandraClusterStatusStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newCassandraCluster, ok := new.(*navigator.CassandraCluster)
	if !ok {
		glog.Fatal("received a non-cassandracluster object to update to")
	}
	oldCassandraCluster, ok := old.(*navigator.CassandraCluster)
	if !ok {
		glog.Fatal("received a non-cassandracluster object to update from")
	}
	// Status changes are not allowed to update spec
	newCassandraCluster.Spec = oldCassandraCluster.Spec
}

func (cassandraClusterStatusStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	return nil
}
