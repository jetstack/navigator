load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "uritemplates.go",
        "utils.go",
    ],
    importpath = "gopkg.in/olivere/elastic.v5/uritemplates",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["utils_test.go"],
    importpath = "gopkg.in/olivere/elastic.v5/uritemplates",
    library = ":go_default_library",
)
