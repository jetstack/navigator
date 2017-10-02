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

package pilot

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

func NewStrategy(typer runtime.ObjectTyper) pilotStrategy {
	return pilotStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*navigator.Pilot)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Pilot.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), PilotToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchPilot is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchPilot(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// PilotToSelectableFields returns a field set that represents the object.
func PilotToSelectableFields(obj *navigator.Pilot) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type pilotStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (pilotStrategy) NamespaceScoped() bool {
	return true
}

func (pilotStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
}

func (pilotStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
}

func (pilotStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (pilotStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (pilotStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (pilotStrategy) Canonicalize(obj runtime.Object) {
}

func (pilotStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

// implements interface RESTUpdateStrategy. This implementation validates updates to
// instance.Status updates only and disallows any modifications to the instance.Spec.
type pilotStatusStrategy struct {
	pilotStrategy
}

func (pilotStatusStrategy) PrepareForUpdate(ctx genericapirequest.Context, new, old runtime.Object) {
	newPilot, ok := new.(*navigator.Pilot)
	if !ok {
		glog.Fatal("received a non-pilot object to update to")
	}
	oldPilot, ok := old.(*navigator.Pilot)
	if !ok {
		glog.Fatal("received a non-pilot object to update from")
	}
	// Status changes are not allowed to update spec
	newPilot.Spec = oldPilot.Spec
}

func (pilotStatusStrategy) ValidateUpdate(ctx genericapirequest.Context, new, old runtime.Object) field.ErrorList {
	return nil
}
