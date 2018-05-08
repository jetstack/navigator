#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..

ARGS="${@:--v}"
echo "+++ Running dep ensure with args '${ARGS[@]}'"

dep ensure "${ARGS[@]}"

# Remove symlink from vendor/github.com/coreos/etcd/cmd/etcd to vendor/github.com/coreos/etcd
# as it causes bazel to fail
rm -f vendor/github.com/coreos/etcd/cmd/etcd
# Remove both files and folders named BUILD, as they can cause confusion for
# the bazel build system
echo "+++ Pruning vendor/ directory of old bazel BUILD files"
find 'vendor/' -iname BUILD -exec rm {} \;

echo "+++ Running 'bazel run //:gazelle'"

bazel run //:gazelle
