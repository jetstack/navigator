load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["util.go"],
    importmap = "vendor/github.com/coreos/etcd/clientv3/clientv3util",
    importpath = "github.com/coreos/etcd/clientv3/clientv3util",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/etcd/clientv3:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["example_key_test.go"],
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/coreos/etcd/clientv3:go_default_library"],
)
