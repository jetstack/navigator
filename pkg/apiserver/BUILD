load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["apiserver.go"],
    importpath = "github.com/jetstack/navigator/pkg/apiserver",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator:go_default_library",
        "//pkg/apis/navigator/install:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/client/informers/internalversion:go_default_library",
        "//pkg/registry/navigator/cassandracluster:go_default_library",
        "//pkg/registry/navigator/escluster:go_default_library",
        "//pkg/registry/navigator/pilot:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apimachinery/announced:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apimachinery/registered:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/serializer:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/version:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/registry/rest:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/server:go_default_library",
    ],
)
