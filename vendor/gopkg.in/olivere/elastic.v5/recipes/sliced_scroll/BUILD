load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["sliced_scroll.go"],
    importmap = "vendor/gopkg.in/olivere/elastic.v5/recipes/sliced_scroll",
    importpath = "gopkg.in/olivere/elastic.v5/recipes/sliced_scroll",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/golang.org/x/sync/errgroup:go_default_library",
        "//vendor/gopkg.in/olivere/elastic.v5:go_default_library",
    ],
)

go_binary(
    name = "sliced_scroll",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
