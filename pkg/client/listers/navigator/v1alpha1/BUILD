load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cassandracluster.go",
        "elasticsearchcluster.go",
        "expansion_generated.go",
        "pilot.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/labels:go_default_library",
        "//vendor/k8s.io/client-go/tools/cache:go_default_library",
    ],
)
