load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["controller.go"],
    importpath = "github.com/jetstack/navigator/cmd/controller/app",
    visibility = ["//visibility:public"],
    deps = [
        "//cmd/controller/app/options:go_default_library",
        "//pkg/client/clientset/versioned:go_default_library",
        "//pkg/client/clientset/versioned/scheme:go_default_library",
        "//pkg/client/informers/externalversions:go_default_library",
        "//pkg/controllers:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/client-go/informers:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes/scheme:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes/typed/core/v1:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
        "//vendor/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//vendor/k8s.io/client-go/tools/leaderelection:go_default_library",
        "//vendor/k8s.io/client-go/tools/leaderelection/resourcelock:go_default_library",
        "//vendor/k8s.io/client-go/tools/record:go_default_library",
    ],
)
