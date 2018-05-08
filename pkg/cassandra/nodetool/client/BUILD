load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["client.go"],
    importpath = "github.com/jetstack/navigator/pkg/cassandra/nodetool/client",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/api/version:go_default_library",
        "//vendor/github.com/pborman/uuid:go_default_library",
    ],
)
