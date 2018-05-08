load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "version.go",
        "zz_generated.deepcopy.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/api/version",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/hashicorp/go-version:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["version_test.go"],
    embed = [":go_default_library"],
)
