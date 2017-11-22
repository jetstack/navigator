load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cmp.go",
        "context.go",
        "server.go",
        "unique.go",
    ],
    importpath = "cloud.google.com/go/internal/testutil",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/google/go-cmp/cmp:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/golang.org/x/oauth2/google:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "server_test.go",
        "unique_test.go",
    ],
    importpath = "cloud.google.com/go/internal/testutil",
    library = ":go_default_library",
    deps = ["//vendor/google.golang.org/grpc:go_default_library"],
)
