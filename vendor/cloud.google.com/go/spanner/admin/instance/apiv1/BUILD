load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "instance_admin_client.go",
    ],
    importpath = "cloud.google.com/go/spanner/admin/instance/apiv1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/internal/version:go_default_library",
        "//vendor/cloud.google.com/go/longrunning:go_default_library",
        "//vendor/cloud.google.com/go/longrunning/autogen:go_default_library",
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/iam/v1:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/longrunning:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/spanner/admin/instance/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["mock_test.go"],
    importpath = "cloud.google.com/go/spanner/admin/instance/apiv1",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/empty:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/iam/v1:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/longrunning:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/rpc/status:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/spanner/admin/instance/v1:go_default_library",
        "//vendor/google.golang.org/genproto/protobuf/field_mask:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
        "//vendor/google.golang.org/grpc/status:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["instance_admin_client_example_test.go"],
    importpath = "cloud.google.com/go/spanner/admin/instance/apiv1_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/iam/v1:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/spanner/admin/instance/v1:go_default_library",
    ],
)
