load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "github.com/jetstack/navigator/pkg/client/informers/externalversions/navigator",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/client/informers/externalversions/internalinterfaces:go_default_library",
        "//pkg/client/informers/externalversions/navigator/v1alpha1:go_default_library",
    ],
)
