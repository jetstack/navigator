load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "doc.go",
    ],
    importpath = "gopkg.in/olivere/elastic.v5/config",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["config_test.go"],
    importpath = "gopkg.in/olivere/elastic.v5/config",
    library = ":go_default_library",
)
