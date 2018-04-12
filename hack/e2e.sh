#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Close stdin
exec 0<&-

: ${TEST_PREFIX:=""}

: ${NAVIGATOR_IMAGE_REPOSITORY:="quay.io/jetstack"}
: ${NAVIGATOR_IMAGE_TAG:="build"}
: ${NAVIGATOR_IMAGE_PULLPOLICY:="Never"}

export \
    NAVIGATOR_IMAGE_REPOSITORY \
    NAVIGATOR_IMAGE_TAG \
    NAVIGATOR_IMAGE_PULLPOLICY

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

# Override these variables in order change the repository and pull policy from
# if you've published test images to your own repository.
: ${CHART_VALUES:="${SCRIPT_DIR}/testdata/values.yaml"}
: ${CHART_VALUES_CASSANDRA:="${SCRIPT_DIR}/testdata/values_cassandra.yaml"}

# Save the cluster logs when the script exits (success or failure)
trap "dump_debug_logs ${PWD}/_artifacts/dump_debug_logs" EXIT

helm delete --purge "${RELEASE_NAME}" || true

function debug_navigator_start() {
    kubectl api-versions
    kubectl get pods --all-namespaces
    kubectl describe --namespace "${NAVIGATOR_NAMESPACE}" deploy
    kubectl describe --namespace "${NAVIGATOR_NAMESPACE}"  pod
}

function navigator_install() {
    echo "Installing navigator..."
    helm delete --purge "${RELEASE_NAME}" || true
    kube_delete_namespace_and_wait "${NAVIGATOR_NAMESPACE}"
    kube_create_namespace_with_quota "${NAVIGATOR_NAMESPACE}"
    if helm --debug install \
            --namespace "${NAVIGATOR_NAMESPACE}" \
            --wait \
            --name "${RELEASE_NAME}" contrib/charts/navigator \
            --values ${CHART_VALUES}
    then
        return 0
    fi
    return 1
}

# Retry helm install to work around intermittent API server availability.
# See https://github.com/jetstack/navigator/issues/118
if ! retry navigator_install; then
    debug_navigator_start
    echo "ERROR: Failed to install Navigator"
    exit 1
fi

# Wait for navigator API to be ready
function navigator_ready() {
    local replica_count_controller=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-controller \
                --namespace "${NAVIGATOR_NAMESPACE}" \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_controller}" -eq 0 ]]; then
        return 1
    fi
    local replica_count_apiserver=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-apiserver \
                --namespace "${NAVIGATOR_NAMESPACE}" \
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
    if ! kube_event_exists "${NAVIGATOR_NAMESPACE}" \
         "navigator-controller:Endpoints:Normal:LeaderElection"
    then
        return 1
    fi
    return 0
}

echo "Waiting for Navigator to be ready..."
if ! retry navigator_ready; then
    debug_navigator_start
    echo "ERROR: Timeout waiting for Navigator API"
    exit 1
fi

FAILURE_COUNT=0
TEST_ID="$(date +%s)-${RANDOM}"

function fail_test() {
    FAILURE_COUNT=$(($FAILURE_COUNT+1))
    echo "TEST FAILURE: $1"
}

function test_general() {
    echo "Testing General"
    local namespace="${1}"

    kube_create_namespace_with_quota "${namespace}"

    local invalid_namespace="notarealnamespace"
    echo "Ensuring NamespaceLifecycle admission controller blocks creation of resources in non-existent namespaces"
    if kubectl create --namespace "${invalid_namespace}" -f "${SCRIPT_DIR}/testdata/testpilot.yaml"; then
        fail_test "navigator-apiserver allowed creation of a resource in a namespace that does not exist"
    fi

    echo "Ensuring we can create resources in valid namespaces"
    if ! kubectl create --namespace "${namespace}" -f "${SCRIPT_DIR}/testdata/testpilot.yaml"; then
        fail_test "navigator-apiserver should allow creation of resources in namespaces that exist"
    fi
}

GENERAL_TEST_NS="test-general-${TEST_ID}"
test_general "${GENERAL_TEST_NS}"
if [ "${FAILURE_COUNT}" -gt "0" ]; then
    exit 1
fi
kube_delete_namespace_and_wait "${GENERAL_TEST_NS}"

function test_elasticsearchcluster() {
    local namespace="${1}"
    echo "Testing ElasticsearchCluster"
    kube_create_namespace_with_quota "${namespace}"
    if ! kubectl get esc; then
        fail_test "Failed to use shortname to get ElasticsearchClusters"
    fi
    # Create and delete an ElasticSearchCluster
    if ! kubectl create \
            --namespace "${namespace}" \
            --filename \
            <(envsubst \
                  '$NAVIGATOR_IMAGE_REPOSITORY:$NAVIGATOR_IMAGE_TAG:$NAVIGATOR_IMAGE_PULLPOLICY' \
                  < "${SCRIPT_DIR}/testdata/es-cluster-test.template.yaml")
    then
        fail_test "Failed to create elasticsearchcluster"
    fi
    if ! kubectl get \
            --namespace "${namespace}" \
            ElasticSearchClusters; then
        fail_test "Failed to get elasticsearchclusters"
    fi
    if ! retry kubectl get \
         --namespace "${namespace}" \
         service es-test; then
        fail_test "Navigator controller failed to create elasticsearchcluster service"
    fi
    if ! retry kube_event_exists "${namespace}" \
         "navigator-controller:ElasticsearchCluster:Normal:CreateNodePool"
    then
        fail_test "Navigator controller failed to create CreateNodePool event"
    fi
    # Wait for Elasticsearch pod to enter 'Running' phase
    if ! retry TIMEOUT=300 stdout_equals "Running" kubectl \
        --namespace "${namespace}" \
        get pod \
        "es-test-mixed-0" \
        "-o=go-template={{.status.phase}}"
    then
        fail_test "Elasticsearch pod did not enter 'Running' phase"
    fi
    # A Pilot is elected leader
    if ! retry TIMEOUT=300 kube_event_exists "${namespace}" \
         "generic-pilot:ConfigMap:Normal:LeaderElection"
    then
        fail_test "Elasticsearch pilots did not elect a leader"
    fi
    # Ensure the Pilot updates the document count on the pilot resource
    if ! retry TIMEOUT=300 stdout_gt 0 kubectl \
         --namespace "${namespace}" \
         get pilot \
         "es-test-mixed-0" \
         "-o=go-template={{.status.elasticsearch.documents}}"
    then
        fail_test "Elasticsearch pilot did not update the document count"
    fi
    # Ensure the Pilot reports the overall cluster health back to the API
    if ! retry TIMEOUT=300 stdout_equals "Yellow" kubectl \
        --namespace "${namespace}" \
        get elasticsearchcluster \
        "test" \
        "-o=go-template={{.status.health}}"
    then
        fail_test "Elasticsearch cluster health status should reflect cluster state"
    fi
}

if [[ "test_elasticsearchcluster" = "${TEST_PREFIX}"* ]]; then
    ES_TEST_NS="test-elasticsearchcluster-${TEST_ID}"
    test_elasticsearchcluster "${ES_TEST_NS}"
    if [ "${FAILURE_COUNT}" -gt "0" ]; then
        exit 1
    fi
    kube_delete_namespace_and_wait "${ES_TEST_NS}"
fi

function test_cassandracluster() {
    echo "Testing CassandraCluster"
    local namespace="${1}"

    export CASS_NAME="test"
    export CASS_NODEPOOL1_DATACENTER="region-1"
    export CASS_NODEPOOL1_RACK="zone-a"
    export CASS_NODEPOOL1_NAME="np-${CASS_NODEPOOL1_DATACENTER}-${CASS_NODEPOOL1_RACK}"
    export CASS_REPLICAS=1
    export CASS_CQL_PORT=9042
    export CASS_VERSION="3.11.1"

    kube_create_namespace_with_quota "${namespace}"

    if ! kubectl get \
         --namespace "${namespace}" \
         CassandraClusters; then
        fail_test "Failed to get cassandraclusters"
    fi

    if ! kubectl get \
         --namespace "${namespace}" \
         cassandra; then
         fail_test "Failed to get cassandraclusters short name (cassandra)"
    fi

    if ! kubectl apply \
        --namespace "${namespace}" \
        --filename \
        <(envsubst \
              '$NAVIGATOR_IMAGE_REPOSITORY:$NAVIGATOR_IMAGE_TAG:$NAVIGATOR_IMAGE_PULLPOLICY:$CASS_NAME:$CASS_NODEPOOL1_DATACENTER:$CASS_NODEPOOL1_RACK:$CASS_NODEPOOL1_NAME:$CASS_REPLICAS:$CASS_VERSION' \
              < "${SCRIPT_DIR}/testdata/cass-cluster-test.template.yaml")
    then
        fail_test "Failed to create cassandracluster"
    fi

    # A NodePool is created
    if ! retry TIMEOUT=300 kube_event_exists "${namespace}" \
         "navigator-controller:CassandraCluster:Normal:CreateNodePool"
    then
        fail_test "A CreateNodePool event was not recorded"
    fi

    # A Pilot is elected leader
    if ! retry TIMEOUT=300 kube_event_exists "${namespace}" \
         "generic-pilot:ConfigMap:Normal:LeaderElection"
    then
        fail_test "Cassandra pilots did not elect a leader"
    fi

    if ! retry TIMEOUT=300 \
         stdout_equals "${CASS_VERSION}" \
         kubectl --namespace "${namespace}" \
         get pilots \
         --output 'jsonpath={.items[0].status.cassandra.version}'
    then
        kubectl --namespace "${namespace}" get pilots -o yaml
        fail_test "Pilots failed to report the expected version"
    fi

    # Wait 5 minutes for cassandra to start and listen for CQL queries.
    if ! retry TIMEOUT=300 cql_connect \
         "${namespace}" \
         "cass-${CASS_NAME}-nodes" \
         "${CASS_CQL_PORT}"; then
        fail_test "Navigator controller failed to create cassandracluster service"
    fi

    if ! retry TIMEOUT=300 in_cluster_command \
        "${namespace}" \
        "alpine:3.6" \
        /bin/sh -c "apk add --no-cache curl && curl -vv http://cass-${CASS_NAME}-nodes:8080"; then
        fail_test "Pilot did not start Prometheus metric exporter"
    fi

    # Create a database
    cql_connect \
        "${namespace}" \
        "cass-${CASS_NAME}-nodes" \
        "${CASS_CQL_PORT}" \
        --debug \
        < "${SCRIPT_DIR}/testdata/cassandra_test_database1.cql"

    # Insert a record
    cql_connect \
        "${namespace}" \
        "cass-${CASS_NAME}-nodes" \
        "${CASS_CQL_PORT}" \
        --debug \
        --execute="INSERT INTO space1.testtable1(key, value) VALUES('testkey1', 'testvalue1')"

    # Delete the Cassandra pod and wait for the CQL service to become
    # unavailable (readiness probe fails)

    kubectl --namespace "${namespace}" delete pod "cass-${CASS_NAME}-${CASS_NODEPOOL1_NAME}-0"
    retry \
        not \
        cql_connect \
        "${namespace}" \
        "cass-${CASS_NAME}-nodes" \
        "${CASS_CQL_PORT}" \
        --debug
    # Kill the cassandra process gracefully which allows it to flush its data to disk.
    # kill_cassandra_process \
    #     "${namespace}" \
    #     "cass-${CASS_NAME}-${CASS_NODEPOOL1_NAME}-0" \
    #     "cassandra" \
    #     "SIGTERM"

    # Test that the data is still there after the Cassandra process restarts
    #
    # XXX: The first successful connection to the database should return the testvalue1.
    # I.e. The `stdout_matches` should come before `retry`
    # In practice I'm finding that `kubectl run cqlsh` sometimes succeeds,
    # but does not relay the pod output.
    # Maybe due to https://github.com/kubernetes/kubernetes/issues/27264
    if ! retry TIMEOUT=300 \
         stdout_matches "testvalue1" \
         cql_connect \
         "${namespace}" \
         "cass-${CASS_NAME}-nodes" \
         "${CASS_CQL_PORT}" \
         --debug \
         --execute='SELECT * FROM space1.testtable1'
    then
        fail_test "Cassandra data was lost"
    fi

    # Increment the replica count
    export CASS_REPLICAS=3
    kubectl apply \
        --namespace "${namespace}" \
        --filename \
        <(envsubst \
              '$NAVIGATOR_IMAGE_REPOSITORY:$NAVIGATOR_IMAGE_TAG:$NAVIGATOR_IMAGE_PULLPOLICY:$CASS_NAME:$CASS_NODEPOOL1_DATACENTER:$CASS_NODEPOOL1_RACK:$CASS_NODEPOOL1_NAME:$CASS_REPLICAS:$CASS_VERSION' \
              < "${SCRIPT_DIR}/testdata/cass-cluster-test.template.yaml")

    # The NodePool is scaled out
    if ! retry TIMEOUT=300 kube_event_exists "${namespace}" \
         "navigator-controller:CassandraCluster:Normal:ScaleOut"
    then
        fail_test "A ScaleOut event was not recorded"
    fi

    if ! retry TIMEOUT=300 stdout_equals 3 kubectl \
         --namespace "${namespace}" \
         get cassandracluster \
         "${CASS_NAME}" \
         "-o=jsonpath={ .status.nodePools['${CASS_NODEPOOL1_NAME}'].readyReplicas }"
    then
        fail_test "The extra cassandra nodes did not become ready"
    fi

    # TODO: A better test would be to query the endpoints and check that only
    # the `-0` pods are included. E.g.
    # kubectl -n test-cassandra-1519754828-19864 get ep cass-cassandra-1519754828-19864-cassandra-seeds -o "jsonpath={.subsets[*].addresses[*].hostname}"
    if ! stdout_equals "cass-${CASS_NAME}-${CASS_NODEPOOL1_NAME}-0" \
         kubectl get pods --namespace "${namespace}" \
         --selector=navigator.jetstack.io/cassandra-seed=true \
         --output 'jsonpath={.items[*].metadata.name}'
    then
        fail_test "First cassandra node not marked as seed"
    fi

    if ! retry \
         stdout_matches "testvalue1" \
         cql_connect \
         "${namespace}" \
         "cass-${CASS_NAME}-nodes" \
         "${CASS_CQL_PORT}" \
         --debug \
         --execute='CONSISTENCY ALL; SELECT * FROM space1.testtable1'
    then
        fail_test "Data was not replicated to second node"
    fi

    simulate_unresponsive_cassandra_process \
        "${namespace}" \
        "cass-${CASS_NAME}-${CASS_NODEPOOL1_NAME}-0"

    if ! retry TIMEOUT=600 \
            stdout_matches "testvalue1" \
            cql_connect \
            "${namespace}" \
            "cass-${CASS_NAME}-nodes" \
            "${CASS_CQL_PORT}" \
            --debug \
            --execute='CONSISTENCY ALL; SELECT * FROM space1.testtable1'
    then
        fail_test "Cassandra liveness probe failed to restart dead node"
    fi
}

if [[ "test_cassandracluster" = "${TEST_PREFIX}"* ]]; then
    CASS_TEST_NS="test-cassandra-${TEST_ID}"

    for i in {1..5}; do
        kube_create_pv "${CASS_TEST_NS}-pv${i}" 5Gi default
    done

    test_cassandracluster "${CASS_TEST_NS}"
    if [ "${FAILURE_COUNT}" -gt "0" ]; then
        exit 1
    fi
    kube_delete_namespace_and_wait "${CASS_TEST_NS}"
fi
