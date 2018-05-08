load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "discovery.go",
        "discovery.pb.go",
    ],
    importmap = "vendor/github.com/googleapis/gnostic/discovery",
    importpath = "github.com/googleapis/gnostic/discovery",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/googleapis/gnostic/compiler:go_default_library",
        "//vendor/gopkg.in/yaml.v2:go_default_library",
        "@io_bazel_rules_go//proto/wkt:any_go_proto",
    ],
)
