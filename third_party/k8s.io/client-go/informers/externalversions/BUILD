licenses(["notice"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "factory.go",
        "generic.go",
    ],
    importpath = "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions",
    visibility = ["//visibility:public"],
    deps = [
        "//third_party/k8s.io/client-go/informers/externalversions/apps:go_default_library",
        "//third_party/k8s.io/client-go/informers/externalversions/core:go_default_library",
        "//third_party/k8s.io/client-go/informers/externalversions/internalinterfaces:go_default_library",
        "//third_party/k8s.io/client-go/informers/externalversions/rbac:go_default_library",
        "//vendor/k8s.io/api/apps/v1beta1:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/api/rbac/v1beta1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)
