#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Temporary fix for https://github.com/kubernetes/kubernetes/issues/37598
# This BUILD file doesn't work when vendored into other projects so we remove
# it. This causes gazelle to generate a new BUILD file from scratch
rm vendor/k8s.io/apimachinery/pkg/util/sets/BUILD
# Remove infinitely recursive symlink from coreos/etcd
rm vendor/github.com/coreos/etcd/cmd/etcd

bazel run //:gazelle
