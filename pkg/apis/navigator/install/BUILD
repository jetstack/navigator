load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["install.go"],
    importpath = "github.com/jetstack/navigator/pkg/apis/navigator/install",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apimachinery/announced:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apimachinery/registered:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
    ],
)
