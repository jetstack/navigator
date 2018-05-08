load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["marshal_fail_case.go"],
    importmap = "vendor/github.com/json-iterator/go/output_tests",
    importpath = "github.com/json-iterator/go/output_tests",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/json-iterator/go:go_default_library"],
)

go_binary(
    name = "output_tests",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
