load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "indent.go",
        "wrap.go",
    ],
    importmap = "vendor/github.com/kr/text",
    importpath = "github.com/kr/text",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "indent_test.go",
        "wrap_test.go",
    ],
    embed = [":go_default_library"],
)
