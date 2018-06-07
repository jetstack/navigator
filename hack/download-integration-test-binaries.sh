#!/bin/bash
#
# Download the binaries needed by the sigs.k8s.io/testing_frameworks/integration package.
# etcd, kube-apiserver, kubectl
# XXX: There is already a script to do this:
# * sigs.k8s.io/testing_frameworks/integration/scripts/download-binaries.sh
# But it currently downloads kube-apiserver v1.10.0-alpha.1 which doesn't support ``CustomResourceSubresources``.
# See https://github.com/kubernetes-sigs/testing_frameworks/issues/44

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Close stdin
exec 0<&-

ROOT_DIR="$(git rev-parse --show-toplevel)"

ETCD_VERSION=v3.2.10
ETCD_URL="https://storage.googleapis.com/etcd"

KUBE_VERSION_URL="https://storage.googleapis.com/kubernetes-release/release/stable-1.10.txt"
KUBE_VERSION=$(curl --fail --silent "${KUBE_VERSION_URL}")
KUBE_BIN_URL="https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64"

ASSETS_DIR="${ROOT_DIR}/vendor/sigs.k8s.io/testing_frameworks/integration/assets/bin"

mkdir -p "${ASSETS_DIR}"

curl --fail --silent ${ETCD_URL}/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz | \
    tar --extract --gzip --directory="${ASSETS_DIR}" --strip-components=1 --wildcards '*/etcd'
curl --fail --silent --output "${ASSETS_DIR}/kube-apiserver" "${KUBE_BIN_URL}/kube-apiserver"
curl --fail --silent --output "${ASSETS_DIR}/kubectl" "${KUBE_BIN_URL}/kubectl"

chmod +x ${ASSETS_DIR}/*
