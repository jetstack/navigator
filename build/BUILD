load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_bundle")

DOCKERIZED_BINARIES = {
    "controller": {
        "base": "@alpine//image",
        "target": "//cmd/controller:controller",
    },
    "apiserver": {
        "base": "@alpine//image",
        "target": "//cmd/apiserver:apiserver",
    },
}

DOCKERIZED_PILOTS = {
    "pilot-elasticsearch": {
        "base": "@alpine//image",
        "target": "//cmd/pilot-elasticsearch:pilot-elasticsearch",
    },
    "pilot-cassandra": {
        "base": "//build:cassandra-base",
        "target": "//cmd/pilot-cassandra:pilot-cassandra",
    },
}

container_image(
    name = "cassandra-base",
    base = "@alpine//image",
    files = [
        "//build/pilot-cassandra:deps",
        "@cassandra_prometheus_jmx//file",
        "@cassandra_jolokia//file",
        "@cassandra_seed_provider//file",
    ],
    symlinks = {
        "/jmx_prometheus_javaagent.jar": "jmx_prometheus_javaagent-0.3.0.jar",
        "/jolokia.jar": "jolokia-jvm-1.4.0-agent.jar",
    },
)

[container_image(
    name = binary + "-internal",
    base = meta["base"],
    cmd = ["/" + binary],
    files = [meta["target"]],
    symlinks = {
        "/usr/bin/navigator-" + binary: "/" + binary,
    },
) for binary, meta in DOCKERIZED_BINARIES.items()]

[container_bundle(
    name = binary,
    images = {
        "jetstackexperimental/navigator-%s:{STABLE_DOCKER_TAG}" % binary: binary + "-internal",
    },
    stamp = True,
) for binary in DOCKERIZED_BINARIES.keys()]

[container_image(
    name = binary + "-internal",
    base = meta["base"],
    cmd = ["/" + binary],
    files = [meta["target"]],
    symlinks = {
        "/pilot": "/" + binary,
        "/usr/bin/" + binary: "/" + binary,
    },
) for binary, meta in DOCKERIZED_PILOTS.items()]

[container_bundle(
    name = binary,
    images = {
        "jetstackexperimental/navigator-%s:{STABLE_DOCKER_TAG}" % binary: binary + "-internal",
    },
    stamp = True,
) for binary in DOCKERIZED_PILOTS.keys()]

[genrule(
    name = binary + "_docker_tag",
    srcs = [meta["target"]],
    outs = [binary + ".docker_tag"],
    cmd = "grep ^STABLE_DOCKER_TAG bazel-out/stable-status.txt | awk '{print $$2}' >$@",
    stamp = 1,
) for binary, meta in DOCKERIZED_BINARIES.items()]

[genrule(
    name = binary + "_docker_tag",
    srcs = [meta["target"]],
    outs = [binary + ".docker_tag"],
    cmd = "grep ^STABLE_DOCKER_TAG bazel-out/stable-status.txt | awk '{print $$2}' >$@",
    stamp = 1,
) for binary, meta in DOCKERIZED_PILOTS.items()]
