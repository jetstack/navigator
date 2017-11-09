#!/bin/bash
#
# Install e2e test dependencies on Travis
set -eux

SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
source "${SCRIPT_DIR}/libe2e.sh"

curl -Lo helm.tar.gz \
     https://storage.googleapis.com/kubernetes-helm/helm-v2.6.1-linux-amd64.tar.gz
tar xvf helm.tar.gz
sudo mv linux-amd64/helm /usr/local/bin

curl -Lo kubectl \
     https://storage.googleapis.com/kubernetes-release/release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

curl -Lo minikube \
     https://storage.googleapis.com/minikube/releases/v0.23.0/minikube-linux-amd64
chmod +x minikube
sudo mv minikube /usr/local/bin/

docker run -v /usr/local/bin:/hostbin quay.io/jetstack/ubuntu-nsenter cp /nsenter /hostbin/nsenter

# Create a cluster. We do this as root as we are using the 'docker' driver.
# We enable RBAC on the cluster too, to test the RBAC in Navigators chart
sudo -E CHANGE_MINIKUBE_NONE_USER=true minikube start \
     -v 100 \
     --vm-driver=none \
     --kubernetes-version="$KUBERNETES_VERSION" \
     --extra-config=apiserver.Authorization.Mode=RBAC \
     --bootstrapper=kubeadm

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
