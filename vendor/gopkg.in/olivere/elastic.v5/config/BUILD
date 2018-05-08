load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "doc.go",
    ],
    importmap = "vendor/gopkg.in/olivere/elastic.v5/config",
    importpath = "gopkg.in/olivere/elastic.v5/config",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["config_test.go"],
    embed = [":go_default_library"],
)
