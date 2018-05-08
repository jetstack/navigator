load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "chacha20poly1305.go",
        "chacha20poly1305_amd64.go",
        "chacha20poly1305_amd64.s",
        "chacha20poly1305_generic.go",
        "chacha20poly1305_noasm.go",
    ],
    importmap = "vendor/golang.org/x/crypto/chacha20poly1305",
    importpath = "golang.org/x/crypto/chacha20poly1305",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/crypto/chacha20poly1305/internal/chacha20:go_default_library",
        "//vendor/golang.org/x/crypto/poly1305:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "chacha20poly1305_test.go",
        "chacha20poly1305_vectors_test.go",
    ],
    embed = [":go_default_library"],
)
