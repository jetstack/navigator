#!/bin/bash
set -eux

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"

source "${SCRIPT_DIR}/libe2e.sh"

echo "Installing helm..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: List
items:

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
EOF
helm init --service-account=tiller

echo "Waiting for tiller to be ready..."
retry TIMEOUT=60 helm version

echo "Applying Elasticsearch virtual memory configuration on all nodes..."
# See https://www.elastic.co/guide/en/elasticsearch/reference/current/system-config.html
kubectl apply --filename "${ROOT_DIR}/docs/quick-start/sysctl-daemonset.yaml"
