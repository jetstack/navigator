load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "options.go",
        "pilot.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/pilot/cassandra/v3",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/cassandra/nodetool:go_default_library",
        "//pkg/client/clientset/versioned:go_default_library",
        "//pkg/client/informers/externalversions:go_default_library",
        "//pkg/client/listers/navigator/v1alpha1:go_default_library",
        "//pkg/pilot/genericpilot:go_default_library",
        "//pkg/pilot/genericpilot/hook:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/k8s.io/client-go/tools/clientcmd:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["pilot_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//pkg/api/version:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/cassandra/nodetool:go_default_library",
        "//pkg/cassandra/nodetool/fake:go_default_library",
    ],
)
