load("@io_bazel_rules_go//go:def.bzl", "go_library")

filegroup(
    name = "go_default_library_protos",
    srcs = ["plugin.proto"],
    visibility = ["//visibility:public"],
)

go_library(
    name = "go_default_library",
    srcs = [
        "environment.go",
        "plugin.pb.go",
    ],
    importpath = "github.com/googleapis/gnostic/plugins",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/googleapis/gnostic/OpenAPIv2:go_default_library",
        "//vendor/github.com/googleapis/gnostic/OpenAPIv3:go_default_library",
        "//vendor/github.com/googleapis/gnostic/discovery:go_default_library",
        "//vendor/github.com/googleapis/gnostic/surface:go_default_library",
    ],
)
