#!/bin/bash
set -eux

function not() {
    if ! $@; then
        return 0
    fi
    return 1
}


function retry() {
    local TIMEOUT=60
    local SLEEP=10
    while :
    do
        case "${1}" in
            TIMEOUT=*|SLEEP=*)
                local "${1}"
                shift
                ;;
            *)
                break
                ;;
        esac
    done

    local start_time="$(date +"%s")"
    local end_time="$(($start_time + ${TIMEOUT}))"
    while true; do
        if $@; then
            return 0
        fi
        local current_time="$(date +"%s")"
        local remaining_time="$((end_time - current_time))"
        if [[ "${remaining_time}" -lt 0 ]]; then
            return 1
        fi
        local sleep_time="${SLEEP}"
        if [[ "${remaining_time}" -lt "${SLEEP}" ]]; then
            sleep_time="${remaining_time}"
        fi
        sleep "${sleep_time}"
    done
    return 1
}

function kube_delete_namespace_and_wait() {
    local namespace=$1
    # Delete all the resources in the namespace
    # This is a work around for Kubernetes 1.7 which doesn't support garbage
    # collection of resources owned by third party resources.
    # See https://github.com/kubernetes/kubernetes/issues/44507
    kubectl --namespace "${namespace}" \
            delete \
            services,serviceaccounts,roles,rolebindings,statefulsets,pods \
            --all || true
    # Delete any previous namespace and wait for Kubernetes to finish deleting.
    kubectl delete --now namespace "${namespace}" || true
    if ! retry TIMEOUT=300 not kubectl get namespace ${namespace}; then
        # If the namespace doesn't delete in time, display the remaining
        # resources.
        kubectl cluster-info dump --namespaces "${namespace}" || true
        return 1
    fi
    return 0
}

function kube_event_exists() {
    local namespace="${1}"
    local event="${2}"
    local go_template='{{range .items}}{{.source.component}}:{{.involvedObject.kind}}:{{.type}}:{{.reason}}{{"\n"}}{{end}}'
    if kubectl get \
               --namespace "${namespace}" \
               events \
               --output=go-template="${go_template}" \
            | grep "^${event}$"; then
        return 0
    fi
    return 1
}

function kube_simulate_unresponsive_process() {
    local namespace=$1
    local pod=$2
    local container=$3
    # Send STOP signal to all processes in the root process group
    # https://unix.stackexchange.com/a/149756
    kubectl \
        --namespace="${namespace}" \
        exec "${pod}" --container="${container}" -- \
        kill -SIGSTOP --  -1
}

function stdout_equals() {
    local expected="${1}"
    shift
    local actual=$("${@}")
    if [[ "${expected}" == "${actual}" ]]; then
        return 0
    fi
    return 1
}

function stdout_gt() {
    local expected="${1}"
    shift
    local actual=$("${@}")
    re='^[0-9]+$'
    if ! [[ "${actual}" =~ $re ]]; then
        echo "${actual} is not a number"
        return 1
    fi
    if [[ "${actual}" -gt "${expected}" ]]; then
        return 0
    fi
    return 1
}
