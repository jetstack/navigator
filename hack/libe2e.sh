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
    # Delete any previous namespace and wait for Kubernetes to finish deleting.
    kubectl delete namespace "${namespace}" || true
    retry TIMEOUT=300 not kubectl get namespace ${namespace}
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
