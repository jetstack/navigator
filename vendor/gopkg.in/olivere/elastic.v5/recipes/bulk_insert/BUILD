load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["bulk_insert.go"],
    importmap = "vendor/gopkg.in/olivere/elastic.v5/recipes/bulk_insert",
    importpath = "gopkg.in/olivere/elastic.v5/recipes/bulk_insert",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/golang.org/x/sync/errgroup:go_default_library",
        "//vendor/gopkg.in/olivere/elastic.v5:go_default_library",
    ],
)

go_binary(
    name = "bulk_insert",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
