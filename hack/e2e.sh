#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Close stdin
exec 0<&-

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

    # Create a database
    cql_connect \
        "${namespace}" \
        "cass-${CHART_NAME}-cassandra-cql" \
        9042 \
        --debug \
        < "${SCRIPT_DIR}/testdata/cassandra_test_database1.cql"

    # Insert a record
    cql_connect \
        "${namespace}" \
        "cass-${CHART_NAME}-cassandra-cql" \
        9042 \
        --debug \
        --execute="INSERT INTO space1.testtable1(key, value) VALUES('testkey1', 'testvalue1')"

    # Delete the Cassandra pod and wait for the CQL service to become
    # unavailable (readiness probe fails)

    kubectl --namespace "${namespace}" delete pod "cass-${CHART_NAME}-cassandra-ringnodes-0"
    retry \
        not \
        cql_connect \
        "${namespace}" \
        "cass-${CHART_NAME}-cassandra-cql" \
        9042 \
        --debug
    # Kill the cassandra process gracefully which allows it to flush its data to disk.
    # kill_cassandra_process \
    #     "${namespace}" \
    #     "cass-${CHART_NAME}-cassandra-ringnodes-0" \
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
         "cass-${CHART_NAME}-cassandra-cql" \
         9042 \
         --debug \
         --execute='SELECT * FROM space1.testtable1'
    then
        fail_test "Cassandra data was lost"
    fi

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

    for i in {1..2}; do
        kube_create_pv "${CASS_TEST_NS}-pv${i}" 5Gi default
    done

    test_cassandracluster "${CASS_TEST_NS}"
    dump_debug_logs "${CASS_TEST_NS}"
    if [ "${FAILURE_COUNT}" -gt "0" ]; then
        fail_and_exit "${CASS_TEST_NS}"
    fi
    kube_delete_namespace_and_wait "${CASS_TEST_NS}"
fi
