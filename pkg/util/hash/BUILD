load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["hash.go"],
    importpath = "github.com/jetstack/navigator/pkg/util/hash",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/davecgh/go-spew/spew:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["hash_test.go"],
    importpath = "github.com/jetstack/navigator/pkg/util/hash",
    library = ":go_default_library",
    deps = ["//vendor/github.com/davecgh/go-spew/spew:go_default_library"],
)
