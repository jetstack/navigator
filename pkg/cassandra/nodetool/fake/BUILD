load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["fake.go"],
    importpath = "github.com/jetstack/navigator/pkg/cassandra/nodetool/fake",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/api/version:go_default_library",
        "//pkg/cassandra/nodetool:go_default_library",
        "//pkg/cassandra/nodetool/client:go_default_library",
    ],
)
