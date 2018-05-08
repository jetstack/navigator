load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["leaderelection.go"],
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot/leaderelection",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/tools/leaderelection:go_default_library",
        "//vendor/k8s.io/client-go/tools/leaderelection/resourcelock:go_default_library",
        "//vendor/k8s.io/client-go/tools/record:go_default_library",
    ],
)
