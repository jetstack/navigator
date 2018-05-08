load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["inmem.go"],
    importmap = "vendor/cloud.google.com/go/bigtable/bttest",
    importpath = "cloud.google.com/go/bigtable/bttest",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/admin/v2:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/v2:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/rpc/status:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "@io_bazel_rules_go//proto/wkt:empty_go_proto",
        "@io_bazel_rules_go//proto/wkt:wrappers_go_proto",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "example_test.go",
        "inmem_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/admin/v2:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/v2:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
