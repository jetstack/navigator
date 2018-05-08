load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "hashes.go",
        "keccakf.go",
        "keccakf_amd64.go",
        "keccakf_amd64.s",
        "register.go",
        "sha3.go",
        "shake.go",
        "xor.go",
        "xor_generic.go",
        "xor_unaligned.go",
    ],
    importmap = "vendor/golang.org/x/crypto/sha3",
    importpath = "golang.org/x/crypto/sha3",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["sha3_test.go"],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
)
