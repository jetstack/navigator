load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "main.go",
        "num.go",
        "string.go",
    ],
    importmap = "vendor/github.com/kr/text/cmd/agg",
    importpath = "github.com/kr/text/cmd/agg",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "agg",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
