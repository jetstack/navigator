load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "flowrate.go",
        "io.go",
        "util.go",
    ],
    importmap = "vendor/github.com/mxk/go-flowrate/flowrate",
    importpath = "github.com/mxk/go-flowrate/flowrate",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["io_test.go"],
    embed = [":go_default_library"],
)
