load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cassandracluster.go",
        "elasticsearchcluster.go",
        "interface.go",
        "pilot.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/client/informers/externalversions/navigator/v1alpha1",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/client/clientset/versioned:go_default_library",
        "//pkg/client/informers/externalversions/internalinterfaces:go_default_library",
        "//pkg/client/listers/navigator/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/watch:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)
