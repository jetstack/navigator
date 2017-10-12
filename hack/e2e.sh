#!/bin/bash
set -eux

NAVIGATOR_NAMESPACE="navigator"
USER_NAMESPACE="navigator-e2e-"
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
kube_delete_namespaces_with_prefix "${USER_NAMESPACE}"

echo "Installing navigator..."
helm install \
     --wait \
     --name "${RELEASE_NAME}" contrib/charts/navigator \
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
    if ! kubectl api-versions | grep 'navigator.jetstack.io'; then
        return 1
    fi
    return 0
}

if ! retry TIMEOUT=600 navigator_ready; then
        kubectl get pods --all-namespaces
        kubectl describe deploy
        kubectl describe pod
        kubectl api-versions
        exit 1
fi

function fail_test() {
    echo "$1"
    (
        kubectl get po -o yaml
        kubectl describe po
        kubectl get svc -o yaml
        kubectl describe svc
        kubectl get apiservice -o yaml
        kubectl describe apiservice
        # kubectl logs -c apiserver -l app=navigator,component=apiserver
        # kubectl logs -c controller -l app=navigator,component=controller
    ) > debug.log
    exit 1
}

function test_elasticsearchcluster_shortname() {
    echo "Testing ElasticsearchCluster shortname (esc)"
    if ! retry kubectl get esc; then
        fail_test "Failed to use shortname to get ElasticsearchClusters"
    fi
}

function test_elasticsearchcluster_create() {
    echo "Testing creating ElasticsearchCluster"
    local ns="${USER_NAMESPACE}-es1"
    # Create and delete an ElasticSearchCluster
    if ! kubectl create namespace "${ns}"; then
        fail_test "Failed to create namespace '${ns}'"
    fi
    if ! kubectl create \
            --namespace "${ns}" \
            --filename "${ROOT_DIR}/docs/quick-start/es-cluster-demo.yaml"; then
        fail_test "Failed to create elasticsearchcluster"
    fi
    if ! kubectl get \
            --namespace "${ns}" \
            ElasticSearchClusters; then
        fail_test "Failed to get elasticsearchclusters"
    fi
    if ! kubectl delete \
            --namespace "${ns}" \
            ElasticSearchClusters \
            --all; then
        fail_test "Failed to delete elasticsearchcluster"
    fi
}

test_elasticsearchcluster_shortname
test_elasticsearchcluster_create

function test_cassandracluster_create() {
    echo "Testing creating CassandraCluster"
    local ns="${USER_NAMESPACE}-cassandra1"
    # Create and delete an CassandraCluster
    if ! kubectl create namespace "${ns}"; then
        fail_test "Failed to create namespace '${ns}'"
    fi
    if ! kubectl create \
         --namespace "${ns}" \
         --filename "${ROOT_DIR}/docs/quick-start/cassandra-cluster.yaml"; then
        fail_test "Failed to create CassandraCluster"
    fi
    if ! kubectl get \
         --namespace "${ns}" \
         CassandraClusters; then
        fail_test "Failed to get CassandraClusters"
    fi
    if ! kubectl delete \
         --namespace "${ns}" \
         CassandraClusters \
         --all; then
        fail_test "Failed to delete Cassandraclusters"
    fi
}

test_cassandracluster_create
