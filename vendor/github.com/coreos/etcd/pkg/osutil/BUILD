load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "interrupt_unix.go",
        "interrupt_windows.go",
        "osutil.go",
        "signal.go",
        "signal_linux.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/pkg/osutil",
    importpath = "github.com/coreos/etcd/pkg/osutil",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["osutil_test.go"],
    embed = [":go_default_library"],
)
