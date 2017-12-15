load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["connect.go"],
    importpath = "gopkg.in/olivere/elastic.v5/recipes/connect",
    visibility = ["//visibility:private"],
    deps = ["//vendor/gopkg.in/olivere/elastic.v5:go_default_library"],
)

go_binary(
    name = "connect",
    importpath = "gopkg.in/olivere/elastic.v5/recipes/connect",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
