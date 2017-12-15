load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "factory.go",
        "generic.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/client/informers/internalversion",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator:go_default_library",
        "//pkg/client/clientset/internalversion:go_default_library",
        "//pkg/client/informers/internalversion/internalinterfaces:go_default_library",
        "//pkg/client/informers/internalversion/navigator:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)
