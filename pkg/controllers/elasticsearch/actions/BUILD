load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "create_nodepool.go",
        "create_pilot.go",
        "scale.go",
        "update_version.go",
        "util.go",
    ],
    importpath = "github.com/jetstack/navigator/pkg/controllers/elasticsearch/actions",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/controllers:go_default_library",
        "//pkg/controllers/common:go_default_library",
        "//pkg/controllers/elasticsearch/util:go_default_library",
        "//pkg/util/api:go_default_library",
        "//pkg/util/ptr:go_default_library",
        "//vendor/github.com/coreos/go-semver/semver:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/api/apps/v1beta1:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/resource:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/intstr:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "scale_test.go",
        "update_version_test.go",
        "util_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//internal/test/unit/framework:go_default_library",
        "//internal/test/util:go_default_library",
        "//internal/test/util/generate:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/util/ptr:go_default_library",
        "//vendor/github.com/coreos/go-semver/semver:go_default_library",
        "//vendor/github.com/kr/pretty:go_default_library",
        "//vendor/k8s.io/api/apps/v1beta1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
    ],
)
