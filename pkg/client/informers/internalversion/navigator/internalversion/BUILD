load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cassandracluster.go",
        "elasticsearchcluster.go",
        "interface.go",
        "pilot.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/client/informers/internalversion/navigator/internalversion",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator:go_default_library",
        "//pkg/client/clientset/internalversion:go_default_library",
        "//pkg/client/informers/internalversion/internalinterfaces:go_default_library",
        "//pkg/client/listers/navigator/internalversion:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/watch:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)
