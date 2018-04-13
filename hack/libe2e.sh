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

function kube_create_namespace_with_quota() {
    local namespace=$1
    kubectl create namespace "${namespace}"
    kubectl create quota \
            --namespace "${namespace}" \
            --hard=cpu=16,requests.cpu=16,limits.cpu=16,memory=32G,requests.memory=32G,limits.memory=32G \
            navigator-test-quota
}

function kube_delete_namespace_and_wait() {
    local namespace=$1
    # Delete ESCs and C* clusters in the namespace
    if ! retry kubectl --namespace "${namespace}" \
         delete \
         elasticsearchclusters,cassandraclusters \
         --now \
         --all
    then
        # If multiple attempts to delete resources fails, display the remaining
        # resources.
        return 1
    fi
    # This is a work around for Kubernetes 1.7 which doesn't support garbage
    # collection of resources owned by third party resources.
    # See https://github.com/kubernetes/kubernetes/issues/44507
    if ! retry kubectl --namespace "${namespace}" \
         delete \
         deployments,replicasets,statefulsets,pods \
         --now \
         --all
    then
        # If multiple attempts to delete resources fails, display the remaining
        # resources.
        return 1
    fi
    if ! wait_for_namespace_empty "${namespace}"; \
    then
        return 1
    fi
    return 0
}

# waits for a namespace to contain 0 pods
function wait_for_namespace_empty() {
    local namespace=$1
    if retry TIMEOUT=300 namespace_empty "${namespace}"; then
        return 0
    fi
    return 1
}

function namespace_empty() {
    local namespace=$1
    if stdout_equals "0" kubectl \
        --namespace "${namespace}" \
        get pods \
        --output='go-template={{len .items}}'; then
        return 0
    fi
    return 1
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

function decommission_cassandra_node() {
    local namespace="${1}"
    local pod="${2}"
    kubectl \
        --namespace="${namespace}" \
        exec "${pod}" -- \
        /bin/sh -c 'JVM_OPTS="" exec nodetool decommission'
}

function signal_cassandra_process() {
    local namespace="${1}"
    local pod="${2}"
    local signal="${3}"

    # Send STOP signal to all the cassandra user's processes
    kubectl \
        --namespace="${namespace}" \
        exec "${pod}" -- \
        bash -c "kill -${signal}"' -- $(ps -u cassandra -o pid=) && ps faux'
}

function simulate_unresponsive_cassandra_process() {
    local namespace="${1}"
    local pod="${2}"
    signal_cassandra_process "${namespace}" "${pod}" "SIGSTOP"
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
    grep "${expected}" <<<"${actual}"
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
    local output_dir="${1}"
    echo "Dumping cluster state to ${output_dir}"
    mkdir -p "${output_dir}"
    kubectl cluster-info dump \
            --all-namespaces \
            --output-directory "${output_dir}" || true

    # Some other resources which aren't included in cluster-info dump
    for kind in apiservice cassandraclusters elasticsearchclusters pilots; do
        kubectl get "${kind}" --all-namespaces --output json > "${output_dir}/${kind}.json" || true
    done
    kubectl api-versions > "${output_dir}/api-versions.txt" || true
}

function cql_connect() {
    local namespace="${1}"
    shift

    # Attempt to negotiate a CQL connection.
    # XXX: This uses the standard Cassandra Docker image rather than the
    # gcr.io/google-samples/cassandra image used in the Cassandra chart, becasue
    # cqlsh is missing some dependencies in that image.
    in_cluster_command "${namespace}" "cassandra:latest" /usr/bin/cqlsh "$@"
}

function in_cluster_command() {
    local namespace="${1}"
    shift
    local image="${1}"
    shift
    kubectl \
        run \
        "in-cluster-cmd-${RANDOM}" \
        --namespace="${namespace}" \
        --image="${image}" \
        --restart=Never \
        --rm \
        --stdin=true \
        --attach=true \
        --quiet \
        --limits="cpu=100m,memory=500Mi" \
        --requests="cpu=100m,memory=500Mi" \
        -- \
        "${@}"
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
    local schedulable_nodes
    schedulable_nodes=$(
        kubectl get nodes \
                --output \
                'jsonpath={range $.items[*]}{.metadata.name} {.spec.taints[*].effect}{"\n"}{end}' \
            | grep -v NoSchedule)
    local node_name
    node_name=$(python -c 'import random,sys; print(random.choice(sys.argv[1:]))' $schedulable_nodes)

    local path="hostpath_pvs/${name}/"

    kubectl create --filename - <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: ${name}
  labels:
    purpose: test
  annotations:
        "volume.alpha.kubernetes.io/node-affinity": '{
            "requiredDuringSchedulingIgnoredDuringExecution": {
                "nodeSelectorTerms": [
                    { "matchExpressions": [
                        { "key": "kubernetes.io/hostname",
                          "operator": "In",
                          "values": ["${node_name}"]
                        }
                    ]}
                 ]}
              }'
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: ${capacity}
  local:
    path: "/tmp/${path}"
  storageClassName: ${storage_class}
  persistentVolumeReclaimPolicy: Delete
EOF

    # Run a job (on the target node) to create the host directory.
    kubectl create --namespace kube-system  --filename - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: "navigator-e2e-create-pv-${name}"
  labels:
    purpose: test
spec:
  template:
    spec:
      restartPolicy: "OnFailure"
      nodeSelector:
        kubernetes.io/hostname: "${node_name}"
      containers:
      - name: "mkdir"
        image: "busybox:latest"
        resources:
          limits:
            cpu: "10m"
            memory: "8Mi"
          requests:
            cpu: "10m"
            memory: "8Mi"
        securityContext:
          privileged: true
        command:
        - "/bin/mkdir"
        - "-p"
        - "/HOST_TMP/${path}"
        volumeMounts:
        - mountPath: /HOST_TMP
          name: host-tmp
      volumes:
      - name: host-tmp
        hostPath:
          path: /tmp
EOF
}

function kube_get_pod_uid() {
    local namespace="${1}"
    local pod="${2}"

    kubectl --namespace "${namespace}" get pod \
            "${pod}" \
            --output "jsonpath={ .metadata.uid }" \
        | egrep '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
}

function kube_get_pod_ip() {
    local namespace="${1}"
    local pod="${2}"

    kubectl --namespace "${namespace}" get pod \
            "${pod}" \
            --output "jsonpath={ .status.podIP }" | \
        egrep '^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$'
}

function kube_delete_pod_and_test_for_new_ip() {
    local namespace="${1}"
    local pod="${2}"

    local original_ip
    original_ip=$(retry kube_get_pod_ip "${namespace}" "${pod}")

    local original_uid
    original_uid=$(retry kube_get_pod_uid "${namespace}" "${pod}")

    # Delete the pod without grace period
    kubectl --namespace "${namespace}" delete pod "${pod}" --grace-period=0 --force || true
    # Run another pod immediately, to hopefully take the IP address of the deleted pod
    in_cluster_command "${namespace}" "busybox:latest" "/bin/sleep" "2"
    retry not stdout_equals "${original_uid}" \
          retry kube_get_pod_uid "${namespace}" "${pod}"

    local new_ip
    new_ip=$(retry kube_get_pod_ip "${namespace}" "${pod}")

    [[ "${original_ip}" != "${new_ip}" ]]
}
