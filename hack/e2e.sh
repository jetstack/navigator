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

echo "Installing navigator..."
helm install --wait --name "${RELEASE_NAME}" contrib/charts/navigator \
        --set apiserver.image.pullPolicy=Never \
        --set controller.image.pullPolicy=Never

# Wait for navigator API to be ready
function navigator_ready() {
    if kubectl api-versions | grep 'navigator.jetstack.io'; then
        return 0
    fi
    return 1
}

echo "Waiting for Navigator to be ready..."
if ! retry navigator_ready; then
    (
        kubectl api-versions
        kubectl get pods --all-namespaces
        kubectl describe deploy
        kubectl describe pod
    ) > debug.log
    echo "ERROR: Timeout waiting for Navigator API"
    exit 1
fi

kubectl create namespace "${USER_NAMESPACE}"

FAILURE_COUNT=0

function fail_test() {
    FAILURE_COUNT=$(($FAILURE_COUNT+1))
    echo "TEST FAILURE: $1"
}

function test_elasticsearchcluster() {
    echo "Testing ElasticsearchCluster"
    if ! kubectl get esc; then
        fail_test "Failed to use shortname to get ElasticsearchClusters"
    fi
    # Create and delete an ElasticSearchCluster
    if ! kubectl create \
            --namespace "${USER_NAMESPACE}" \
            --filename "${ROOT_DIR}/docs/quick-start/es-cluster-demo.yaml"; then
        fail_test "Failed to create elasticsearchcluster"
    fi
    if ! kubectl get \
            --namespace "${USER_NAMESPACE}" \
            ElasticSearchClusters; then
        fail_test "Failed to get elasticsearchclusters"
    fi
    if ! kubectl delete \
            --namespace "${USER_NAMESPACE}" \
            ElasticSearchClusters \
            --all; then
        fail_test "Failed to delete elasticsearchcluster"
    fi
}

test_elasticsearchcluster

if [[ "${FAILURE_COUNT}" -gt 0 ]]; then
    (
        kubectl get po -o yaml
        kubectl describe po
        kubectl get svc -o yaml
        kubectl describe svc
        kubectl get apiservice -o yaml
        kubectl describe apiservice
        kubectl logs -c apiserver -l app=navigator,component=apiserver
        kubectl logs -c controller -l app=navigator,component=controller
    ) > debug.log
fi

exit $FAILURE_COUNT
