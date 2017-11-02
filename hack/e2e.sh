#!/bin/bash
set -eux

NAVIGATOR_NAMESPACE="navigator"
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

echo "Installing navigator..."
helm install --wait --name "${RELEASE_NAME}" contrib/charts/navigator \
        --set apiserver.image.pullPolicy=Never \
        --set controller.image.pullPolicy=Never

# Wait for navigator API to be ready
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
    if ! kubectl api-versions | grep 'navigator.jetstack.io'; then
        return 1
    fi
    # Even after the API appears in api-versions, it takes a short time for API
    # server to recognise navigator API types.
    if ! kubectl get esc; then
        return 1
    fi
    if ! kube_event_exists "kube-system" \
         "navigator-controller:Endpoints:Normal:LeaderElection"
    then
        return 1
    fi
    return 0
}

echo "Waiting for Navigator to be ready..."
if ! retry navigator_ready; then
    kubectl api-versions
    kubectl get pods --all-namespaces
    kubectl describe deploy
    kubectl describe pod
    echo "ERROR: Timeout waiting for Navigator API"
    exit 1
fi

TEST_ID="${RANDOM}"
FAILURE_COUNT=0

function fail_test() {
    FAILURE_COUNT=$(($FAILURE_COUNT+1))
    echo "TEST FAILURE: $1"
}

function test_elasticsearchcluster_success() {
    echo "Testing ElasticsearchCluster success path"
    local FAILURE_COUNT=0
    local NAMESPACE="${TEST_ID}-test-elasticsearchcluster-success"
    kubectl create namespace "${NAMESPACE}"

    if ! kubectl get esc; then
        fail_test "Failed to use shortname to get ElasticsearchClusters"
    fi
    # Create and delete an ElasticSearchCluster
    if ! kubectl create \
            --namespace "${NAMESPACE}" \
            --filename "${ROOT_DIR}/docs/quick-start/es-cluster-demo.yaml"; then
        fail_test "Failed to create elasticsearchcluster"
    fi
    if ! kubectl get \
            --namespace "${NAMESPACE}" \
            ElasticSearchClusters; then
        fail_test "Failed to get elasticsearchclusters"
    fi
    if ! retry kubectl get \
         --namespace "${NAMESPACE}" \
         service es-demo; then
        fail_test "Navigator controller failed to create elasticsearchcluster service"
    fi
    if ! retry kube_event_exists "${NAMESPACE}" \
         "navigator-controller:ElasticsearchCluster:Normal:SuccessSync"
    then
        fail_test "Navigator controller failed to create SuccessSync event"
    fi
    if ! kubectl delete \
            --namespace "${NAMESPACE}" \
            ElasticSearchClusters \
            --all; then
        fail_test "Failed to delete elasticsearchcluster"
    fi

    if kubectl get --namespace "${NAMESPACE}" events \
            | grep 'Warning'; then
        fail_test "unexpected warnings found"
    fi


    if [[ "${FAILURE_COUNT}" -eq 0 ]]; then
        kubectl delete namespace "${NAMESPACE}"
    fi
    return $FAILURE_COUNT
}

if ! test_elasticsearchcluster_success; then
    fail_test "test_elasticsearchcluster_success"
fi


function test_elasticsearchcluster_failure() {
    echo "Testing ElasticsearchCluster failure path"
    local FAILURE_COUNT=0
    local NAMESPACE="${TEST_ID}-test-elasticsearchcluster-failure"
    kubectl create namespace "${NAMESPACE}"

    # Create a clashing servicaccount name to trigger a controller sync failure
    kubectl create --namespace "${NAMESPACE}" \
            serviceaccount es-demo
    if ! kubectl create \
            --namespace "${NAMESPACE}" \
            --filename "${ROOT_DIR}/docs/quick-start/es-cluster-demo.yaml"; then
        fail_test "Failed to create elasticsearchcluster"
    fi
    if ! retry kube_event_exists "${NAMESPACE}" \
         "navigator-controller:ElasticsearchCluster:Warning:ErrorSync"
    then
        fail_test "Navigator controller failed to create ErrorSync event"
    fi
    if [[ "${FAILURE_COUNT}" -eq 0 ]]; then
        kubectl delete namespace "${NAMESPACE}"
    fi
    return $FAILURE_COUNT
}

if ! test_elasticsearchcluster_failure; then
    fail_test "test_elasticsearchcluster_failure"
fi

function test_logs() {
    if kubectl logs deployments/nav-e2e-navigator-controller \
            | grep '^E' \
            | grep -v 'reflector.go:205'; then
        fail_test "Unexpected errors in controller logs"
    fi
}

test_logs

if [[ "${FAILURE_COUNT}" -gt 0 ]]; then
    kubectl get po -o yaml
    kubectl describe po
    kubectl get svc -o yaml
    kubectl describe svc
    kubectl get apiservice -o yaml
    kubectl describe apiservice
    kubectl logs -c apiserver -l app=navigator,component=apiserver
    kubectl logs -c controller -l app=navigator,component=controller
fi

exit $FAILURE_COUNT
