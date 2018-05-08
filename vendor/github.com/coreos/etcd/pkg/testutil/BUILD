load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "assert.go",
        "leak.go",
        "pauseable_handler.go",
        "recorder.go",
        "testutil.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/pkg/testutil",
    importpath = "github.com/coreos/etcd/pkg/testutil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["leak_test.go"],
    embed = [":go_default_library"],
)
