load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "interval_tree.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/pkg/adt",
    importpath = "github.com/coreos/etcd/pkg/adt",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "example_test.go",
        "interval_tree_test.go",
    ],
    embed = [":go_default_library"],
)
