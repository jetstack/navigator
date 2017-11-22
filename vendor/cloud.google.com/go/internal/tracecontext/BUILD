load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["tracecontext.go"],
    importpath = "cloud.google.com/go/internal/tracecontext",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["tracecontext_test.go"],
    importpath = "cloud.google.com/go/internal/tracecontext",
    library = ":go_default_library",
    deps = ["//vendor/cloud.google.com/go/internal/testutil:go_default_library"],
)
