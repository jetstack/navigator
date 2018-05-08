load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "signal.go",
        "signal_posix.go",
        "signal_windows.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot/signals",
    visibility = ["//visibility:public"],
)
