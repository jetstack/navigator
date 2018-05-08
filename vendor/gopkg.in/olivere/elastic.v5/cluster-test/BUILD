load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["cluster-test.go"],
    importmap = "vendor/gopkg.in/olivere/elastic.v5/cluster-test",
    importpath = "gopkg.in/olivere/elastic.v5/cluster-test",
    visibility = ["//visibility:private"],
    deps = ["//vendor/gopkg.in/olivere/elastic.v5:go_default_library"],
)

go_binary(
    name = "cluster-test",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
