#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

echo "Running 'bazel run //:gazelle'"

bazel run //:gazelle
