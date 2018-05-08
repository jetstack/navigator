load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["mc.go"],
    importmap = "vendor/github.com/kr/text/mc",
    importpath = "github.com/kr/text/mc",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/kr/pty:go_default_library",
        "//vendor/github.com/kr/text/colwriter:go_default_library",
    ],
)

go_binary(
    name = "mc",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
