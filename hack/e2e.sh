#!/bin/bash
set -eux


: ${TEST_PREFIX:=""}

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
: ${CHART_VALUES_CASSANDRA:="${SCRIPT_DIR}/testdata/values_cassandra.yaml"}

FAILURE_COUNT=0
TEST_ID="$(date +%s)-${RANDOM}"

function fail_test() {
    FAILURE_COUNT=$(($FAILURE_COUNT+1))
    echo "TEST FAILURE: $1"
}

# function test_elasticsearchcluster() {
#     local namespace="${1}"
#     echo "Testing ElasticsearchCluster"
#     kubectl create namespace "${namespace}"
#     if ! kubectl get esc; then
#         fail_test "Failed to use shortname to get ElasticsearchClusters"
#     fi
#     # Create and delete an ElasticSearchCluster
#     if ! kubectl create \
#             --namespace "${namespace}" \
#             --filename \
#             <(envsubst \
#                   '$NAVIGATOR_IMAGE_REPOSITORY:$NAVIGATOR_IMAGE_TAG:$NAVIGATOR_IMAGE_PULLPOLICY' \
#                   < "${SCRIPT_DIR}/testdata/es-cluster-test.template.yaml")
#     then
#         fail_test "Failed to create elasticsearchcluster"
#     fi
#     if ! kubectl get \
#             --namespace "${namespace}" \
#             ElasticSearchClusters; then
#         fail_test "Failed to get elasticsearchclusters"
#     fi
#     if ! retry kubectl get \
#          --namespace "${namespace}" \
#          service es-test; then
#         fail_test "Navigator controller failed to create elasticsearchcluster service"
#     fi
#     if ! retry kube_event_exists "${namespace}" \
#          "navigator-controller:ElasticsearchCluster:Normal:CreateNodePool"
#     then
#         fail_test "Navigator controller failed to create CreateNodePool event"
#     fi
#     # Wait for Elasticsearch pod to enter 'Running' phase
#     if ! retry TIMEOUT=300 stdout_equals "Running" kubectl \
#         --namespace "${namespace}" \
#         get pod \
#         "es-test-mixed-0" \
#         "-o=go-template={{.status.phase}}"
#     then
#         fail_test "Elasticsearch pod did not enter 'Running' phase"
#     fi
#     # A Pilot is elected leader
#     if ! retry TIMEOUT=300 kube_event_exists "${namespace}" \
#          "generic-pilot:ConfigMap:Normal:LeaderElection"
#     then
#         fail_test "Elasticsearch pilots did not elect a leader"
#     fi
#     # Ensure the Pilot updates the document count on the pilot resource
#     if ! retry TIMEOUT=300 stdout_gt 0 kubectl \
#          --namespace "${namespace}" \
#          get pilot \
#          "es-test-mixed-0" \
#          "-o=go-template={{.status.elasticsearch.documents}}"
#     then
#         fail_test "Elasticsearch pilot did not update the document count"
#     fi
#     # Ensure the Pilot reports the overall cluster health back to the API
#     if ! retry TIMEOUT=300 stdout_equals "Yellow" kubectl \
#         --namespace "${namespace}" \
#         get elasticsearchcluster \
#         "test" \
#         "-o=go-template={{.status.health}}"
#     then
#         fail_test "Elasticsearch cluster health status should reflect cluster state"
#     fi
# }

# if [[ "test_elasticsearchcluster" = "${TEST_PREFIX}"* ]]; then
#     ES_TEST_NS="test-elasticsearchcluster-${TEST_ID}"
#     test_elasticsearchcluster "${ES_TEST_NS}"
#     if [ "${FAILURE_COUNT}" -gt "0" ]; then
#         fail_and_exit "${ES_TEST_NS}"
#     fi
#     kube_delete_namespace_and_wait "${ES_TEST_NS}"
# fi

function cql_connect() {
    local namespace="${1}"
    local host="${2}"
    local port="${3}"
    # Attempt to negotiate a CQL connection.
    # No queries are performed.
    # stdin=false (the default) ensures that cqlsh does not go into interactive
    # mode.
    kubectl \
        run \
        "cql-responding-${RANDOM}" \
        --namespace="${namespace}" \
        --command=true \
        --image=cassandra:3 \
        --restart=Never \
        --rm \
        --stdin=false \
        --attach=true \
        -- \
        /usr/bin/cqlsh --debug "${host}" "${port}"
}

function test_cassandracluster() {
    echo "Testing CassandraCluster"
    local namespace="${1}"
    local CHART_NAME="cassandra-${TEST_ID}"

    kubectl create namespace "${namespace}"

    if ! kubectl get \
         --namespace "${namespace}" \
         CassandraClusters; then
        fail_test "Failed to get cassandraclusters"
    fi

    helm install \
         --debug \
         --wait \
         --name "${CHART_NAME}" \
         --namespace "${namespace}" \
         contrib/charts/cassandra \
         --values "${CHART_VALUES_CASSANDRA}" \
         --set replicaCount=1

    # A Pilot is elected leader
    if ! retry TIMEOUT=300 kube_event_exists "${namespace}" \
         "generic-pilot:ConfigMap:Normal:LeaderElection"
    then
        fail_test "Cassandra pilots did not elect a leader"
    fi

    # Wait 5 minutes for cassandra to start and listen for CQL queries.
    if ! retry TIMEOUT=300 cql_connect \
         "${namespace}" \
         "cass-${CHART_NAME}-cassandra-cql" \
         9042; then
        fail_test "Navigator controller failed to create cassandracluster service"
    fi

    # TODO Fail test if there are unexpected cassandra errors.
    kubectl log \
            --namespace "${namespace}" \
            "statefulset/cass-${CHART_NAME}-cassandra-ringnodes"

    # Change the CQL port
    helm --debug upgrade \
         "${CHART_NAME}" \
         contrib/charts/cassandra \
         --values "${CHART_VALUES_CASSANDRA}" \
         --set replicaCount=1 \
         --set cqlPort=9043

    # Wait 60s for cassandra CQL port to change
    if ! retry TIMEOUT=60 cql_connect \
         "${namespace}" \
         "cass-${CHART_NAME}-cassandra-cql" \
         9043; then
        fail_test "Navigator controller failed to update cassandracluster service"
    fi

    # Increment the replica count
    helm --debug upgrade \
         "${CHART_NAME}" \
         contrib/charts/cassandra \
         --values "${CHART_VALUES_CASSANDRA}" \
         --set cqlPort=9043 \
         --set replicaCount=2

    if ! retry TIMEOUT=300 stdout_equals 2 kubectl \
         --namespace "${namespace}" \
         get statefulsets \
         "cass-${CHART_NAME}-cassandra-ringnodes" \
         "-o=go-template={{.status.readyReplicas}}"
    then
        fail_test "Second cassandra node did not become ready"
    fi

    simulate_unresponsive_cassandra_process \
        "${namespace}" \
        "cass-${CHART_NAME}-cassandra-ringnodes-0" \
        "cassandra"

    if ! retry cql_connect \
         "${namespace}" \
         "cass-${CHART_NAME}-cassandra-cql" \
         9043; then
        fail_test "Cassandra readiness probe failed to bypass dead node"
    fi
}

if [[ "test_cassandracluster" = "${TEST_PREFIX}"* ]]; then
    CASS_TEST_NS="test-cassandra-${TEST_ID}"
    test_cassandracluster "${CASS_TEST_NS}"
    if [ "${FAILURE_COUNT}" -gt "0" ]; then
        fail_and_exit "${CASS_TEST_NS}"
    fi
    kube_delete_namespace_and_wait "${CASS_TEST_NS}"
fi
