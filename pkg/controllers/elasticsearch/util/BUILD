load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "configmap.go",
        "nodepool.go",
        "role.go",
        "service.go",
        "serviceaccount.go",
        "util.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/controllers/elasticsearch/util",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/labels:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/selection:go_default_library",
    ],
)
