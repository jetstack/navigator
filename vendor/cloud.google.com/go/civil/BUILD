load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["civil.go"],
    importmap = "vendor/cloud.google.com/go/civil",
    importpath = "cloud.google.com/go/civil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["civil_test.go"],
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/google/go-cmp/cmp:go_default_library"],
)
