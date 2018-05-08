load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "blake2b.go",
        "blake2bAVX2_amd64.go",
        "blake2bAVX2_amd64.s",
        "blake2b_amd64.go",
        "blake2b_amd64.s",
        "blake2b_generic.go",
        "blake2b_ref.go",
        "blake2x.go",
        "register.go",
    ],
    importmap = "vendor/golang.org/x/crypto/blake2b",
    importpath = "golang.org/x/crypto/blake2b",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["blake2b_test.go"],
    embed = [":go_default_library"],
)
