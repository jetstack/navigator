#!/bin/bash
set -eux

NAVIGATOR_NAMESPACE="navigator"
USER_NAMESPACE="navigator-e2e-database1"
RELEASE_NAME="nav-e2e"

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
CONFIG_DIR=$(mktemp -d -t navigator-e2e.XXXXXXXXX)
mkdir -p $CONFIG_DIR
CERT_DIR="$CONFIG_DIR/certs"
mkdir -p $CERT_DIR
TEST_DIR="$CONFIG_DIR/tmp"
mkdir -p $TEST_DIR

source "${SCRIPT_DIR}/libe2e.sh"

helm delete --purge "${RELEASE_NAME}" || true
kube_delete_namespace_and_wait "${USER_NAMESPACE}"

echo "Waiting up to 5 minutes for Kubernetes to be ready..."
retry TIMEOUT=600 kubectl get nodes

echo "Installing helm..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: List
items:

### Fix kube-dns ###
- apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRoleBinding
  metadata:
    name: system:kube-dns
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: system:kube-dns
  subjects:
  - kind: ServiceAccount
    name: default
    namespace: kube-system

### Tiller ###
# Create a ServiceAccount for tiller to use
- apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: tiller
    namespace: kube-system
# Bind tiller to the cluster-admin role
- apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRoleBinding
  metadata:
    name: "tiller"
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: "cluster-admin"
  subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: tiller
    namespace: kube-system
### Generic ###
# Create a ClusterRole to work with ElasticsearchCluster resources
- apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRole
  metadata:
    name: navigator:authenticated
  # this rule defined on the role for specifically the
  # namespace-lifecycle admission-controller
  rules:
  - apiGroups: ["navigator.jetstack.io"]
    resources: ["elasticsearchclusters"]
    verbs:     ["get", "list", "watch", "create", "update", "delete"]
- apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRoleBinding
  metadata:
    name: "navigator:authenticated"
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: navigator:authenticated
  subjects:
  - kind: Group
    name: system:authenticated
    apiGroup: rbac.authorization.k8s.io
  - kind: Group
    name: system:unauthenticated
    apiGroup: rbac.authorization.k8s.io
EOF
helm init --service-account=tiller
