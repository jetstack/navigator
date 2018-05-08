load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["nodetool.go"],
    importpath = "github.com/jetstack/navigator/pkg/cassandra/nodetool",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/api/version:go_default_library",
        "//pkg/cassandra/nodetool/client:go_default_library",
        "//vendor/github.com/pborman/uuid:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["nodetool_test.go"],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
    deps = [
        "//pkg/cassandra/nodetool/client:go_default_library",
        "//pkg/cassandra/nodetool/fake:go_default_library",
        "//vendor/github.com/pborman/uuid:go_default_library",
    ],
)
