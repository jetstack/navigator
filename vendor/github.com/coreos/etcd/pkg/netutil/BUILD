load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "isolate_linux.go",
        "isolate_stub.go",
        "netutil.go",
        "routes.go",
        "routes_linux.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/pkg/netutil",
    importpath = "github.com/coreos/etcd/pkg/netutil",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
    ] + select({
        "@io_bazel_rules_go//go/platform:linux": [
            "//vendor/github.com/coreos/etcd/pkg/cpuutil:go_default_library",
        ],
        "//conditions:default": [],
    }),
)

go_test(
    name = "go_default_test",
    srcs = [
        "netutil_test.go",
        "routes_linux_test.go",
    ],
    embed = [":go_default_library"],
)
