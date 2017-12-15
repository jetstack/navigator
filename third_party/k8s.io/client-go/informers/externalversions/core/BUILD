licenses(["notice"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions/core",
    visibility = ["//visibility:public"],
    deps = [
        "//third_party/k8s.io/client-go/informers/externalversions/core/v1:go_default_library",
        "//third_party/k8s.io/client-go/informers/externalversions/internalinterfaces:go_default_library",
    ],
)
