load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "github.com/jetstack/navigator/pkg/client/informers/internalversion/navigator",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/client/informers/internalversion/internalinterfaces:go_default_library",
        "//pkg/client/informers/internalversion/navigator/internalversion:go_default_library",
    ],
)
