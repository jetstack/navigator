#!/bin/bash
#
# Install e2e test dependencies on Travis
set -eux

SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
source "${SCRIPT_DIR}/libe2e.sh"

# Create a cluster. We do this as root as we are using the 'docker' driver.
# The kubeadm bootstrapper enables RBAC by default.
minikube start \
     -v 100 \
     --vm-driver=kvm \
     --kubernetes-version="$KUBERNETES_VERSION" \
     --bootstrapper=kubeadm \
     --profile="$HOSTNAME"

echo "Waiting up to 5 minutes for Kubernetes to be ready..."
if ! retry TIMEOUT=300 kubectl get nodes; then
    minikube logs
    echo "ERROR: Timeout waiting for Minikube to be ready"
    exit 1
fi

# Fix kube-dns RBAC issues.
# Allow kube-dns and other kube-system services full access to the API.
# See:
# * https://github.com/kubernetes/minikube/issues/1734
# * https://github.com/kubernetes/minikube/issues/1722
kubectl create clusterrolebinding cluster-admin:kube-system \
        --clusterrole=cluster-admin \
        --serviceaccount=kube-system:default
