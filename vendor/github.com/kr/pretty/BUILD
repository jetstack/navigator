load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "diff.go",
        "formatter.go",
        "pretty.go",
        "zero.go",
    ],
    importmap = "vendor/github.com/kr/pretty",
    importpath = "github.com/kr/pretty",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/kr/text:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "diff_test.go",
        "example_test.go",
        "formatter_test.go",
    ],
    embed = [":go_default_library"],
)
