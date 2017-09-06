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

echo "Waiting for tiller to be ready..."
retry TIMEOUT=60 helm version

echo "Installing navigator..."
helm install --wait --name "${RELEASE_NAME}" contrib/charts/navigator \
        --set apiserver.image.pullPolicy=Never \
        --set controller.image.pullPolicy=Never

# Wait for navigator pods to be running
function navigator_ready() {
    local replica_count_controller=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-controller \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_controller}" -eq 0 ]]; then
        return 1
    fi
    local replica_count_apiserver=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-apiserver \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_apiserver}" -eq 0 ]]; then
        return 1
    fi
    return 0
}

if ! retry navigator_ready; then
        kubectl get pods --all-namespaces
        kubectl describe deploy
        kubectl describe pod
        exit 1
fi

function apiversion_ready() {
    local apiversion_navigator_length=$(
        kubectl api-versions | grep 'navigator.jetstack.io' | wc -l
    )
    if [[ "${apiversion_navigator_length}" -lt 1 ]]; then
        return 1
    fi
    sleep 15
    return 0
}

echo "Waiting for navigator API version to be registered"
retry TIMEOUT=30 apiversion_ready

kubectl api-versions

echo "Testing ElasticsearchCluster shortname (esc)"
if ! kubectl get esc; then
    echo "Failed to use shortname to get ElasticsearchClusters"
    exit 1
fi

echo "Testing creating ElasticsearchCluster"
# Create and delete an ElasticSearchCluster
kubectl create namespace "${USER_NAMESPACE}"
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
