load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["helloworld.pb.go"],
    importmap = "vendor/cloud.google.com/go/trace/testdata/helloworld",
    importpath = "cloud.google.com/go/trace/testdata/helloworld",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
