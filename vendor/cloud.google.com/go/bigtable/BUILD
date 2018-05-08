load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "admin.go",
        "bigtable.go",
        "doc.go",
        "filter.go",
        "gc.go",
        "reader.go",
    ],
    importmap = "vendor/cloud.google.com/go/bigtable",
    importpath = "cloud.google.com/go/bigtable",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable/internal/gax:go_default_library",
        "//vendor/cloud.google.com/go/bigtable/internal/option:go_default_library",
        "//vendor/cloud.google.com/go/longrunning:go_default_library",
        "//vendor/cloud.google.com/go/longrunning/autogen:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport/grpc:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/admin/v2:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/v2:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
        "@io_bazel_rules_go//proto/wkt:duration_go_proto",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "admin_test.go",
        "bigtable_test.go",
        "export_test.go",
        "gc_test.go",
        "reader_test.go",
        "retry_test.go",
    ],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable/bttest:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/admin/v2:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/bigtable/v2:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/rpc/status:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "@io_bazel_rules_go//proto/wkt:wrappers_go_proto",
    ],
)
