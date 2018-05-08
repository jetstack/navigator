load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["elasticsearch.go"],
    importpath = "github.com/jetstack/navigator/pkg/util/api",
    visibility = ["//visibility:public"],
    deps = ["//pkg/apis/navigator/v1alpha1:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["elasticsearch_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/util/ptr:go_default_library",
    ],
)
