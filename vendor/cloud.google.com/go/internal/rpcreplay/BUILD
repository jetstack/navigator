load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "rpcreplay.go",
    ],
    importpath = "cloud.google.com/go/internal/rpcreplay",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
    deps = [
        "//vendor/cloud.google.com/go/internal/rpcreplay/proto/rpcreplay:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/any:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/rpc/status:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
        "//vendor/google.golang.org/grpc/status:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fake_test.go",
        "rpcreplay_test.go",
    ],
    importpath = "cloud.google.com/go/internal/rpcreplay",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/rpcreplay/proto/intstore:go_default_library",
        "//vendor/cloud.google.com/go/internal/rpcreplay/proto/rpcreplay:go_default_library",
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/status:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "cloud.google.com/go/internal/rpcreplay_test",
)
