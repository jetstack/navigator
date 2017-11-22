load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/googleapis/gnostic/plugins/gnostic-summary",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/googleapis/gnostic/OpenAPIv2:go_default_library",
        "//vendor/github.com/googleapis/gnostic/OpenAPIv3:go_default_library",
        "//vendor/github.com/googleapis/gnostic/plugins:go_default_library",
        "//vendor/github.com/googleapis/gnostic/printer:go_default_library",
    ],
)

go_binary(
    name = "gnostic-summary",
    importpath = "github.com/googleapis/gnostic/plugins/gnostic-summary",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
