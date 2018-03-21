#!/bin/bash
set -eux

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"

source "${SCRIPT_DIR}/libe2e.sh"

NAVIGATOR_NAMESPACE="navigator"
RELEASE_NAME="nav-e2e"
# Override these variables in order change the repository and pull policy from
# if you've published test images to your own repository.
: ${CHART_VALUES:="${SCRIPT_DIR}/testdata/values.yaml"}

echo "Installing helm..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: List
items:

### Tiller ###
# Create a ServiceAccount for tiller to use
- apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: tiller
    namespace: kube-system
# Bind tiller to the cluster-admin role
- apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRoleBinding
  metadata:
    name: "tiller"
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: "cluster-admin"
  subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: tiller
    namespace: kube-system
EOF
helm init --service-account=tiller

echo "Waiting for tiller to be ready..."
retry TIMEOUT=60 helm version

function debug_navigator_start() {
    kubectl api-versions
    kubectl get pods --all-namespaces
    kubectl describe --namespace "${NAVIGATOR_NAMESPACE}" deploy
    kubectl describe --namespace "${NAVIGATOR_NAMESPACE}"  pod
}

function navigator_install() {
    echo "Installing navigator..."
    helm delete --purge "${RELEASE_NAME}" || true
    kube_create_namespace_with_quota "${NAVIGATOR_NAMESPACE}"
    if helm --debug install \
            --namespace "${NAVIGATOR_NAMESPACE}" \
            --wait \
            --name "${RELEASE_NAME}" contrib/charts/navigator \
            --values ${CHART_VALUES}
    then
        return 0
    fi
    return 1
}

# Retry helm install to work around intermittent API server availability.
# See https://github.com/jetstack/navigator/issues/118
if ! retry navigator_install; then
    debug_navigator_start
    echo "ERROR: Failed to install Navigator"
    exit 1
fi

# Wait for navigator API to be ready
function navigator_ready() {
    local replica_count_controller=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-controller \
                --namespace "${NAVIGATOR_NAMESPACE}" \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_controller}" -eq 0 ]]; then
        return 1
    fi
    local replica_count_apiserver=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-apiserver \
                --namespace "${NAVIGATOR_NAMESPACE}" \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_apiserver}" -eq 0 ]]; then
        return 1
    fi
    if ! kubectl api-versions | grep 'navigator.jetstack.io'; then
        return 1
    fi
    # Even after the API appears in api-versions, it takes a short time for API
    # server to recognise navigator API types.
    if ! kubectl get esc; then
        return 1
    fi
    if ! kube_event_exists "${NAVIGATOR_NAMESPACE}" \
         "navigator-controller:Endpoints:Normal:LeaderElection"
    then
        return 1
    fi
    return 0
}

echo "Waiting for Navigator to be ready..."
if ! retry navigator_ready; then
    debug_navigator_start
    echo "ERROR: Timeout waiting for Navigator API"
    exit 1
fi

echo "Applying Elasticsearch virtual memory configuration on all nodes..."
# See https://www.elastic.co/guide/en/elasticsearch/reference/current/system-config.html
kubectl apply --filename "${ROOT_DIR}/docs/quick-start/sysctl-daemonset.yaml"
