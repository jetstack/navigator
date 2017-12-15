load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "controller.go",
        "genericpilot.go",
        "healthz.go",
        "options.go",
        "signal.go",
        "util.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "signal_posix.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "signal_posix.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "signal_windows.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "github.com/jetstack/navigator/pkg/pilot/genericpilot",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/client/clientset/versioned:go_default_library",
        "//pkg/client/clientset/versioned/scheme:go_default_library",
        "//pkg/client/informers/externalversions:go_default_library",
        "//pkg/client/listers/navigator/v1alpha1:go_default_library",
        "//pkg/pilot/genericpilot/hook:go_default_library",
        "//pkg/pilot/genericpilot/probe:go_default_library",
        "//pkg/pilot/genericpilot/process:go_default_library",
        "//pkg/pilot/genericpilot/scheduler:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/wait:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes/scheme:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes/typed/core/v1:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/k8s.io/client-go/tools/record:go_default_library",
        "//vendor/k8s.io/client-go/util/workqueue:go_default_library",
    ],
)
