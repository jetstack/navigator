load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["secretbox.go"],
    importmap = "vendor/k8s.io/apiserver/pkg/storage/value/encrypt/secretbox",
    importpath = "k8s.io/apiserver/pkg/storage/value/encrypt/secretbox",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/crypto/nacl/secretbox:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/storage/value:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["secretbox_test.go"],
    embed = [":go_default_library"],
    deps = ["//vendor/k8s.io/apiserver/pkg/storage/value:go_default_library"],
)
