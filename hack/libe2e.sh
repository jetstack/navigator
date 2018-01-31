function not() {
    if ! "$@"; then
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
    local start_time
    start_time="$(date +"%s")"
    local end_time
    end_time="$(($start_time + ${TIMEOUT}))"
    until "${@}"
    do
        local exit_code="${?}"
        local current_time="$(date +"%s")"
        local remaining_time="$((end_time - current_time))"
        if [[ "${remaining_time}" -le 0 ]]; then
            return "${exit_code}"
        fi
        local sleep_time="${SLEEP}"
        if [[ "${remaining_time}" -lt "${SLEEP}" ]]; then
            sleep_time="${remaining_time}"
        fi
        sleep "${sleep_time}"
    done
}

function kube_delete_namespace_and_wait() {
    local namespace=$1
    # Delete all the resources in the namespace
    # This is a work around for Kubernetes 1.7 which doesn't support garbage
    # collection of resources owned by third party resources.
    # See https://github.com/kubernetes/kubernetes/issues/44507
    if ! retry kubectl --namespace "${namespace}" \
         delete \
         services,serviceaccounts,roles,rolebindings,statefulsets,pods \
         --now \
         --all
    then
        # If multiple attempts to delete resources fails, display the remaining
        # resources.
        return 1
    fi
    # Delete any previous namespace and wait for Kubernetes to finish deleting.
    kubectl delete --now namespace "${namespace}" || true
    if ! retry TIMEOUT=300 not kubectl get namespace ${namespace}; then
        # If the namespace doesn't delete in time, display the remaining
        # resources.
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

function simulate_unresponsice_cassandra_process() {
    local namespace=$1
    local pod=$2
    local container=$3
    # Decommission causes cassandra to stop accepting CQL connections.
    kubectl \
        --namespace="${namespace}" \
        exec "${pod}" --container="${container}" -- \
        nodetool decommission
}

function signal_cassandra_process() {
    local namespace=$1
    local pod=$2
    local container=$3
    local signal=$4

    # Send STOP signal to all the cassandra user's processes
    kubectl \
        --namespace="${namespace}" \
        exec "${pod}" --container="${container}" -- \
        bash -c "kill -${signal}"' -- $(ps -u cassandra -o pid=) && ps faux'
}

function stdout_equals() {
    local expected="${1}"
    shift
    local actual
    actual=$("${@}")
    if [[ "${expected}" == "${actual}" ]]; then
        return 0
    fi
    return 1
}

function stdout_matches() {
    local expected="${1}"
    shift
    local actual
    actual=$("${@}")
    grep --quiet "${expected}" <<<"${actual}"
}

function stdout_gt() {
    local expected="${1}"
    shift
    local actual
    actual=$("${@}")
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

function dump_debug_logs() {
    local namespace="${1}"
    local output_dir="$(pwd)/_artifacts/${namespace}"

    echo "Dumping cluster state to ${output_dir}"
    mkdir -p "${output_dir}"
    kubectl cluster-info dump \
            --namespaces "${namespace}" \
            --output-directory "${output_dir}" || true
}

function fail_and_exit() {
    local namespace="${1}"

    kubectl api-versions
    kubectl get apiservice -o yaml

    dump_debug_logs "${namespace}"

    exit 1
}

function cql_connect() {
    local namespace="${1}"
    shift

    # Attempt to negotiate a CQL connection.
    # XXX: This uses the standard Cassandra Docker image rather than the
    # gcr.io/google-samples/cassandra image used in the Cassandra chart, becasue
    # cqlsh is missing some dependencies in that image.
    kubectl \
        run \
        "cql-connect-${RANDOM}" \
        --namespace="${namespace}" \
        --command=true \
        --image=cassandra:latest \
        --restart=Never \
        --rm \
        --stdin=true \
        --attach=true \
        --quiet \
        -- \
        /usr/bin/cqlsh "$@"
}

function kill_cassandra_process() {
    local namespace=$1
    local pod=$2
    local container=$3
    local signal=$4
    local current_restart_count
    current_restart_count=$(
        kubectl --namespace "${namespace}" get pod "${pod}" -o \
            'jsonpath={.status.containerStatuses[?(@.name=="cassandra")].restartCount}')

    signal_cassandra_process \
        "${namespace}" \
        "${pod}" \
        "${container}" \
        "${signal}"

    retry \
        stdout_gt "${current_restart_count}" \
        kubectl --namespace "${namespace}" get pod "${pod}" -o \
        'jsonpath={.status.containerStatuses[?(@.name=="cassandra")].restartCount}'
}


function kube_create_pv() {
    local name="${1}"
    local capacity="${2}"
    local storage_class="${3}"

    kubectl create --filename - <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${name}
  labels:
    purpose: test
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: ${capacity}
  hostPath:
    path: /data/${name}/
  storageClassName: ${storage_class}
  persistentVolumeReclaimPolicy: Delete
EOF

}
