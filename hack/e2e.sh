#!/bin/bash
set -eux

NAVIGATOR_NAMESPACE="navigator"

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
CONFIG_DIR=$(mktemp --directory --tmpdir navigator-e2e.XXXXXXXXX)
mkdir --parents $CONFIG_DIR
CERT_DIR="$CONFIG_DIR/certs"
mkdir --parents $CERT_DIR
TEST_DIR="$CONFIG_DIR/tmp"
mkdir --parents $TEST_DIR


source "${SCRIPT_DIR}/libe2e.sh"

# Install navigator

kube_delete_namespace_and_wait "${NAVIGATOR_NAMESPACE}"

kubectl create --filename ${ROOT_DIR}/docs/quick-start/deployment-navigator.yaml

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
