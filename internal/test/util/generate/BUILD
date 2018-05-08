load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["generate.go"],
    importpath = "github.com/jetstack/navigator/internal/test/util/generate",
    visibility = ["//:__subpackages__"],
    deps = [
        "//pkg/api/version:go_default_library",
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//vendor/github.com/coreos/go-semver/semver:go_default_library",
        "//vendor/k8s.io/api/apps/v1beta1:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
    ],
)
