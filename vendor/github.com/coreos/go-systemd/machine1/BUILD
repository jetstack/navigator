load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["dbus.go"],
    importmap = "vendor/github.com/coreos/go-systemd/machine1",
    importpath = "github.com/coreos/go-systemd/machine1",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/godbus/dbus:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["dbus_test.go"],
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/coreos/go-systemd/dbus:go_default_library"],
)
