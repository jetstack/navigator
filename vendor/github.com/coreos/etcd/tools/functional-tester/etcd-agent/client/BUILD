load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "client.go",
        "doc.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/tools/functional-tester/etcd-agent/client",
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-agent/client",
    visibility = ["//visibility:public"],
)
