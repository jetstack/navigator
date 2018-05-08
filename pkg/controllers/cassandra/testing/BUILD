load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "gen.go",
        "testing.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/controllers/cassandra/testing",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/api/version:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/util/ptr:go_default_library",
    ],
)
