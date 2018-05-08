load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "fields.go",
        "fold.go",
    ],
    importmap = "vendor/cloud.google.com/go/internal/fields",
    importpath = "cloud.google.com/go/internal/fields",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
    deps = ["//vendor/cloud.google.com/go/internal/atomiccache:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fields_test.go",
        "fold_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/github.com/google/go-cmp/cmp:go_default_library",
    ],
)
