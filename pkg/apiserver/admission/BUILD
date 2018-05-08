load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["initializer.go"],
    importpath = "github.com/jetstack/navigator/pkg/apiserver/admission",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/client/clientset/internalversion:go_default_library",
        "//pkg/client/informers/internalversion:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/admission:go_default_library",
        "//vendor/k8s.io/client-go/informers:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
    ],
)
