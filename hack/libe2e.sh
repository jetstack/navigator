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

function kube_namespaces_with_prefix() {
    local namespace_prefix=$1
    kubectl get namespaces \
            --output "jsonpath={.items[*].metadata.name}" \
        | xargs --no-run-if-empty --max-args 1 \
        | grep "${namespace_prefix}"
}

function kube_namespaces_exist() {
    local namespace_prefix=$1
    local matching_namespaces=$(kube_namespaces_with_prefix "${namespace_prefix}")
    if test -z "${matching_namespaces}"; then
        return 1
    fi
    return 0
}

function kube_delete_namespaces_with_prefix() {
    local namespace_prefix=$1
    # Delete any previous namespace and wait for Kubernetes to finish deleting.
    kube_namespaces_with_prefix "$namespace_prefix" \
        | xargs --no-run-if-empty kubectl delete namespace
    retry not kube_namespaces_exist "${namespace_prefix}"
}
