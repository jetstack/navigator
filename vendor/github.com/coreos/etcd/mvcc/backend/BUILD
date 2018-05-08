load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "backend.go",
        "batch_tx.go",
        "config_default.go",
        "config_linux.go",
        "config_windows.go",
        "doc.go",
        "metrics.go",
        "read_tx.go",
        "tx_buffer.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/mvcc/backend",
    importpath = "github.com/coreos/etcd/mvcc/backend",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/boltdb/bolt:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "backend_bench_test.go",
        "backend_test.go",
        "batch_tx_test.go",
    ],
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/boltdb/bolt:go_default_library"],
)
