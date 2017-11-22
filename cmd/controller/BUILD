load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "main.go",
        "start.go",
    ],
    importpath = "github.com/jetstack/navigator/cmd/controller",
    visibility = ["//visibility:private"],
    deps = [
        "//cmd/controller/app:go_default_library",
        "//cmd/controller/app/options:go_default_library",
        "//pkg/controllers/elasticsearch:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/server:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/util/logs:go_default_library",
        "//vendor/k8s.io/client-go/plugin/pkg/client/auth:go_default_library",
    ],
)

go_binary(
    name = "controller",
    importpath = "github.com/jetstack/navigator/cmd/controller",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
