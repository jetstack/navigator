load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["loadtest.go"],
    importpath = "cloud.google.com/go/pubsub/loadtest/cmd",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/pubsub/loadtest:go_default_library",
        "//vendor/cloud.google.com/go/pubsub/loadtest/pb:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_binary(
    name = "cmd",
    importpath = "cloud.google.com/go/pubsub/loadtest/cmd",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
