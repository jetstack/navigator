load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["rpcreplay.pb.go"],
    importmap = "vendor/cloud.google.com/go/internal/rpcreplay/proto/rpcreplay",
    importpath = "cloud.google.com/go/internal/rpcreplay/proto/rpcreplay",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "@io_bazel_rules_go//proto/wkt:any_go_proto",
    ],
)
