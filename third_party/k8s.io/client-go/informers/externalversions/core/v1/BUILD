licenses(["notice"])

load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "componentstatus.go",
        "configmap.go",
        "endpoints.go",
        "event.go",
        "interface.go",
        "limitrange.go",
        "namespace.go",
        "node.go",
        "persistentvolume.go",
        "persistentvolumeclaim.go",
        "pod.go",
        "podtemplate.go",
        "replicationcontroller.go",
        "resourcequota.go",
        "secret.go",
        "service.go",
        "serviceaccount.go",
    ],
    importpath = "github.com/jetstack/navigator/third_party/k8s.io/client-go/informers/externalversions/core/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//third_party/k8s.io/client-go/informers/externalversions/internalinterfaces:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/watch:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/listers/core/v1:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)
