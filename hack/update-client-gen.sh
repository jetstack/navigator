#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

${CODEGEN_PKG}/generate-internal-groups.sh "deepcopy,defaulter,client,informer,lister" \
  github.com/jetstack/navigator/pkg/client github.com/jetstack/navigator/pkg/apis github.com/jetstack/navigator/pkg/apis \
  navigator:v1alpha1 \
  --output-base "${GOPATH}/src/" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt

# We have to run conversion-gen separately so we can set the --extra-peer-dirs
# flag to not include k8s.io/kubernetes packages (https://github.com/kubernetes/kubernetes/issues/54301)

${CODEGEN_PKG}/generate-internal-groups.sh "conversion" \
  github.com/jetstack/navigator/pkg/client github.com/jetstack/navigator/pkg/apis github.com/jetstack/navigator/pkg/apis \
  navigator:v1alpha1 \
  --output-base "${GOPATH}/src/" \
  --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt \
  --extra-peer-dirs="k8s.io/api/core/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime"

${GOPATH}/bin/informer-gen \
           --output-base "${GOPATH}/src/" \
           --input-dirs "k8s.io/api/core/v1" \
           --input-dirs "k8s.io/api/apps/v1beta1" \
           --input-dirs "k8s.io/api/rbac/v1beta1" \
           --versioned-clientset-package "k8s.io/client-go/kubernetes" \
           --listers-package "k8s.io/client-go/listers" \
           --output-package "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers" \
           --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt
