load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "process.go",
        "signals.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot/process",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/glog:go_default_library"],
)
