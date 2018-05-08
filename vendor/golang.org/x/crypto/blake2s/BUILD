load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "blake2s.go",
        "blake2s_386.go",
        "blake2s_386.s",
        "blake2s_amd64.go",
        "blake2s_amd64.s",
        "blake2s_generic.go",
        "blake2s_ref.go",
        "blake2x.go",
        "register.go",
    ],
    importmap = "vendor/golang.org/x/crypto/blake2s",
    importpath = "golang.org/x/crypto/blake2s",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["blake2s_test.go"],
    embed = [":go_default_library"],
)
