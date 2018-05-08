load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "mockclient.go",
        "mockserver.go",
    ],
    importmap = "vendor/cloud.google.com/go/spanner/internal/testutil",
    importpath = "cloud.google.com/go/spanner/internal/testutil",
    visibility = ["//vendor/cloud.google.com/go/spanner:__subpackages__"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/spanner/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "@io_bazel_rules_go//proto/wkt:empty_go_proto",
        "@io_bazel_rules_go//proto/wkt:struct_go_proto",
        "@io_bazel_rules_go//proto/wkt:timestamp_go_proto",
    ],
)
