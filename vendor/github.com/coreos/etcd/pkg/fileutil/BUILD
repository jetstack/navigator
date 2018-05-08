load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "dir_unix.go",
        "dir_windows.go",
        "fileutil.go",
        "lock.go",
        "lock_flock.go",
        "lock_linux.go",
        "lock_plan9.go",
        "lock_solaris.go",
        "lock_unix.go",
        "lock_windows.go",
        "preallocate.go",
        "preallocate_darwin.go",
        "preallocate_unix.go",
        "preallocate_unsupported.go",
        "purge.go",
        "sync.go",
        "sync_darwin.go",
        "sync_linux.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/pkg/fileutil",
    importpath = "github.com/coreos/etcd/pkg/fileutil",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fileutil_test.go",
        "lock_test.go",
        "preallocate_test.go",
        "purge_test.go",
    ],
    embed = [":go_default_library"],
)
