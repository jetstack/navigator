#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

CASS_TEMPLATE="${SCRIPT_ROOT}/hack/testdata/cass-cluster-test.template.yaml"
ES_TEMPLATE="${SCRIPT_ROOT}/hack/testdata/es-cluster-test.template.yaml"

EXAMPLES_DIR="${SCRIPT_ROOT}/docs/quick-start/"
CASS_EXAMPLE="${EXAMPLES_DIR}/cassandra-cluster.yaml"
ES_EXAMPLE="${EXAMPLES_DIR}/es-cluster-demo.yaml"

export NAVIGATOR_IMAGE_REPOSITORY="jetstackexperimental"
export NAVIGATOR_IMAGE_TAG="canary"
export NAVIGATOR_IMAGE_PULLPOLICY="IfNotPresent"

export CASS_NAME="demo"
export CASS_REPLICAS=3
export CASS_CQL_PORT=9042
export CASS_VERSION="3.11.1"

envsubst < ${CASS_TEMPLATE} > ${CASS_EXAMPLE}
