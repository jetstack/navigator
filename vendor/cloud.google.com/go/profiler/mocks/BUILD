load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["mock_profiler_client.go"],
    importpath = "cloud.google.com/go/profiler/mocks",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/mock/gomock:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/devtools/cloudprofiler/v2:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
