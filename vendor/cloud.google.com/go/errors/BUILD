load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["errors.go"],
    importmap = "vendor/cloud.google.com/go/errors",
    importpath = "cloud.google.com/go/errors",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/errorreporting/apiv1beta1:go_default_library",
        "//vendor/cloud.google.com/go/internal/version:go_default_library",
        "//vendor/cloud.google.com/go/logging:go_default_library",
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/devtools/clouderrorreporting/v1beta1:go_default_library",
        "@io_bazel_rules_go//proto/wkt:timestamp_go_proto",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "error_logging_test.go",
        "errors_test.go",
        "stack_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//vendor/cloud.google.com/go/logging:go_default_library",
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/devtools/clouderrorreporting/v1beta1:go_default_library",
    ],
)
