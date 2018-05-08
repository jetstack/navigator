load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["atomiccache.go"],
    importmap = "vendor/cloud.google.com/go/internal/atomiccache",
    importpath = "cloud.google.com/go/internal/atomiccache",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["atomiccache_test.go"],
    embed = [":go_default_library"],
)
