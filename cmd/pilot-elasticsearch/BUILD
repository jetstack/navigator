load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "main.go",
        "pilot.go",
    ],
    importpath = "github.com/jetstack/navigator/cmd/pilot-elasticsearch",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/pilot/elasticsearch/v5:go_default_library",
        "//pkg/pilot/genericpilot:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/k8s.io/apiserver/pkg/util/logs:go_default_library",
        "//vendor/k8s.io/client-go/plugin/pkg/client/auth:go_default_library",
    ],
)

go_binary(
    name = "pilot-elasticsearch",
    importpath = "github.com/jetstack/navigator/cmd/pilot-elasticsearch",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
