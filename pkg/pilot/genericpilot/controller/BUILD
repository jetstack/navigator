load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "controller.go",
        "genericpilot.go",
        "util.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot/controller",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/client/clientset/versioned:go_default_library",
        "//pkg/client/informers/externalversions/navigator/v1alpha1:go_default_library",
        "//pkg/client/listers/navigator/v1alpha1:go_default_library",
        "//pkg/controllers:go_default_library",
        "//pkg/pilot/genericpilot/controller/scheduler:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/wait:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/k8s.io/client-go/util/workqueue:go_default_library",
    ],
)
