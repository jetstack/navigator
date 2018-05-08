load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "create_nodepool.go",
        "scaleout.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/controllers/cassandra/actions",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/controllers:go_default_library",
        "//pkg/controllers/cassandra/nodepool:go_default_library",
        "//pkg/util/ptr:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/pkg/errors:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "create_nodepool_test.go",
        "scaleout_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//internal/test/unit/framework:go_default_library",
        "//internal/test/util/generate:go_default_library",
        "//pkg/util/ptr:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
    ],
)
