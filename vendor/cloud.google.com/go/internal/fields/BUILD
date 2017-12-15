load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "fields.go",
        "fold.go",
    ],
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
    importpath = "cloud.google.com/go/internal/fields",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/github.com/google/go-cmp/cmp:go_default_library",
    ],
)
