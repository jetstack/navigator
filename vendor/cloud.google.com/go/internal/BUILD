load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["retry.go"],
    importmap = "vendor/cloud.google.com/go/internal",
    importpath = "cloud.google.com/go/internal",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["retry_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
