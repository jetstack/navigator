load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["scheduler.go"],
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot/controller/scheduler",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["scheduler_test.go"],
    embed = [":go_default_library"],
)
