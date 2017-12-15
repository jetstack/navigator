load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["loadtest.go"],
    importpath = "cloud.google.com/go/pubsub/loadtest",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/pubsub:go_default_library",
        "//vendor/cloud.google.com/go/pubsub/loadtest/pb:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/time/rate:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["benchmark_test.go"],
    importpath = "cloud.google.com/go/pubsub/loadtest",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/cloud.google.com/go/pubsub:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport/grpc:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/pubsub/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
