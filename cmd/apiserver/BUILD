load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "main.go",
        "start.go",
    ],
    importpath = "github.com/jetstack/navigator/cmd/apiserver",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/apis/navigator/v1alpha1:go_default_library",
        "//pkg/apiserver:go_default_library",
        "//pkg/client/clientset/internalversion:go_default_library",
        "//pkg/client/informers/internalversion:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/server:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/server/options:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/util/logs:go_default_library",
    ],
)

go_binary(
    name = "apiserver",
    importpath = "github.com/jetstack/navigator/cmd/apiserver",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
