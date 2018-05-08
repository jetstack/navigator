load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["contains.go"],
    importpath = "github.com/jetstack/navigator/internal/test/util",
    visibility = ["//:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["contains_test.go"],
    embed = [":go_default_library"],
)
