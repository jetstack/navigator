load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "constraint.go",
        "version.go",
        "version_collection.go",
    ],
    importmap = "vendor/github.com/hashicorp/go-version",
    importpath = "github.com/hashicorp/go-version",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "constraint_test.go",
        "version_collection_test.go",
        "version_test.go",
    ],
    embed = [":go_default_library"],
)
