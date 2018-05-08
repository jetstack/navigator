load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["hook.go"],
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot/hook",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
    ],
)
