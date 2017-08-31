#!/bin/bash
set -eux

NAVIGATOR_NAMESPACE="navigator"
USER_NAMESPACE="navigator-e2e-database1"

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
CONFIG_DIR=$(mktemp --directory --tmpdir navigator-e2e.XXXXXXXXX)
mkdir --parents $CONFIG_DIR
CERT_DIR="$CONFIG_DIR/certs"
mkdir --parents $CERT_DIR
TEST_DIR="$CONFIG_DIR/tmp"
mkdir --parents $TEST_DIR

source "${SCRIPT_DIR}/libe2e.sh"

echo "Waiting up to 5 minutes for Kubernetes to be ready..."
retry TIMEOUT=600 kubectl get nodes

kube_delete_namespace_and_wait "${NAVIGATOR_NAMESPACE}"
kube_delete_namespace_and_wait "${USER_NAMESPACE}"
kubectl delete ThirdPartyResources --all

# Install navigator
kubectl create \
        --filename "${ROOT_DIR}/docs/quick-start/deployment-navigator.yaml"

# Wait for navigator pods to be running
function navigator_ready() {
    local replica_count=$(
        kubectl get deployment navigator \
                --namespace navigator \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count}" -eq 1 ]]; then
        return 0
    fi
    return 1
}

retry navigator_ready
kubectl get pods --namespace navigator
kubectl logs --namespace navigator -l app=navigator

# Create and delete an ElasticSearchCluster
kubectl create \
        --namespace "${USER_NAMESPACE}" \
        --filename "${ROOT_DIR}/docs/quick-start/es-cluster-demo.yaml"
kubectl get \
        --namespace "${USER_NAMESPACE}" \
        ElasticSearchClusters
kubectl delete \
        --namespace "${USER_NAMESPACE}" \
        ElasticSearchClusters \
        --all
