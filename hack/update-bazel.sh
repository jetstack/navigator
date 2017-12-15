#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Temporary fix for https://github.com/kubernetes/kubernetes/issues/37598
# This BUILD file doesn't work when vendored into other projects so we remove
# it. This causes gazelle to generate a new BUILD file from scratch
if [ -d vendor/k8s.io/apimachinery/pkg/util/sets/BUILD ]; then rm vendor/k8s.io/apimachinery/pkg/util/sets/BUILD; fi
# Remove infinitely recursive symlink from coreos/etcd
if [ -d vendor/github.com/coreos/etcd/cmd/etcd ]; then rm vendor/github.com/coreos/etcd/cmd/etcd; fi
# Remove bash scripts named 'build', as this causes problems on case
# insensitive file systems (e.g. OSX)
rm -f \
    vendor/github.com/coreos/etcd/build \
    vendor/github.com/coreos/etcd/tools/functional-tester/build \
    vendor/github.com/coreos/etcd/contrib/systemd/etcd2-backup-coreos/build

bazel run //:gazelle
