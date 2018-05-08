load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["control.go"],
    importpath = "github.com/jetstack/navigator/pkg/controllers/cassandra/seedlabeller",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/controllers/cassandra/service:go_default_library",
        "//pkg/controllers/cassandra/util:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/api/apps/v1beta1:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/listers/apps/v1beta1:go_default_library",
        "//vendor/k8s.io/client-go/listers/core/v1:go_default_library",
        "//vendor/k8s.io/client-go/tools/record:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["control_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//internal/test/unit/framework:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/controllers:go_default_library",
        "//pkg/controllers/cassandra/nodepool:go_default_library",
        "//pkg/controllers/cassandra/service:go_default_library",
        "//pkg/controllers/cassandra/testing:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
    ],
)
