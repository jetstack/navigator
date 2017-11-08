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

FAILURE_COUNT=0
TEST_ID="$(date +%s)-${RANDOM}"

function fail_test() {
    FAILURE_COUNT=$(($FAILURE_COUNT+1))
    echo "TEST FAILURE: $1"
}

function test_elasticsearchcluster() {
    echo "Testing ElasticsearchCluster"
    local USER_NAMESPACE="test-elasticsearchcluster-${TEST_ID}"
    kubectl create namespace "${USER_NAMESPACE}"
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
    if ! retry kubectl get \
         --namespace "${USER_NAMESPACE}" \
         service es-demo; then
        fail_test "Navigator controller failed to create elasticsearchcluster service"
    fi
    if ! retry kube_event_exists "${USER_NAMESPACE}" \
         "navigator-controller:ElasticsearchCluster:Normal:SuccessSync"
    then
        fail_test "Navigator controller failed to create SuccessSync event"
    fi
    if ! kubectl delete \
            --namespace "${USER_NAMESPACE}" \
            ElasticSearchClusters \
            --all; then
        fail_test "Failed to delete elasticsearchcluster"
    fi
}

test_elasticsearchcluster

function kube_service_responding() {
    local namespace="${1}"
    local host="${2}"
    local port="${3}"
    kubectl run \
            --namespace="${namespace}" \
            "kube-service-responding-${RANDOM}" \
            --stdin=true\
            --rm \
            --restart=Never \
            --image alpine \
            -- \
            nc -w 5 -v "${host}" "${port}" <<< ping
}

function test_cassandracluster() {
    echo "Testing CassandraCluster"
    local USER_NAMESPACE="test-cassandracluster-${TEST_ID}"
    local CHART_NAME="cassandra-${TEST_ID}"
    kubectl create namespace "${USER_NAMESPACE}"

    if ! kubectl get \
         --namespace "${USER_NAMESPACE}" \
         CassandraClusters; then
        fail_test "Failed to get cassandraclusters"
    fi

    helm install \
         --wait \
         --name "${CHART_NAME}" \
         --namespace "${USER_NAMESPACE}" \
         contrib/charts/cassandra \
         --set replicaCount=1 \

    # Wait 5 minutes for cassandra to start and listen for CQL queries.
    if ! retry TIMEOUT=300 kube_service_responding \
         "${USER_NAMESPACE}" \
         "cass-${CHART_NAME}-cassandra" \
         9042; then
        fail_test "Navigator controller failed to create cassandracluster service"
    fi

    # TODO Fail test if there are unexpected cassandra errors.
    kubectl log \
            --namespace "${USER_NAMESPACE}" \
            "statefulset/cass-${CHART_NAME}-cassandra-ringnodes"
}

test_cassandracluster

function ignore_expected_errors() {
    # Ignored failures to list navigator API objects when the controller starts
    # before the API server has started and registered its self. E.g.
    # E1103 14:58:06.819858       1 reflector.go:205] github.com/jetstack/navigator/pkg/client/informers/externalversions/factory.go:68: Failed to list *v1alpha1.Pilot: the server could not find the requested resource (get pilots.navigator.jetstack.io)
    egrep --invert-match 'Failed to list \*v1alpha1\.\w+:\s+the server could not find the requested resource\s+\(get \w+\.navigator\.jetstack\.io\)$'
}

function test_logged_errors() {
    if kubectl logs deployments/nav-e2e-navigator-controller \
            | egrep '^E' \
            | ignore_expected_errors
    then
        fail_test "Unexpected errors in controller logs"
    fi
}

test_logged_errors

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
