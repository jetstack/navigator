load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "grpc.go",
        "http.go",
        "sampling.go",
        "trace.go",
    ],
    importpath = "cloud.google.com/go/trace",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/internal/tracecontext:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/time/rate:go_default_library",
        "//vendor/google.golang.org/api/cloudtrace/v1:go_default_library",
        "//vendor/google.golang.org/api/gensupport:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/support/bundler:go_default_library",
        "//vendor/google.golang.org/api/transport/http:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "grpc_test.go",
        "http_test.go",
        "trace_test.go",
    ],
    importpath = "cloud.google.com/go/trace",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/datastore:go_default_library",
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/cloud.google.com/go/storage:go_default_library",
        "//vendor/cloud.google.com/go/trace/testdata/helloworld:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/cloudtrace/v1:go_default_library",
        "//vendor/google.golang.org/api/compute/v1:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/datastore/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["httpexample_test.go"],
    importpath = "cloud.google.com/go/trace_test",
    deps = [":go_default_library"],
)
