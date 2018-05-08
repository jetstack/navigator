load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "eventlog_formatter.go",
        "log.go",
        "syslog_formatter.go",
    ],
    importmap = "vendor/github.com/prometheus/common/log",
    importpath = "github.com/prometheus/common/log",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/sirupsen/logrus:go_default_library",
        "//vendor/gopkg.in/alecthomas/kingpin.v2:go_default_library",
    ] + select({
        "@io_bazel_rules_go//go/platform:windows": [
            "//vendor/golang.org/x/sys/windows/svc/eventlog:go_default_library",
        ],
        "//conditions:default": [],
    }),
)

go_test(
    name = "go_default_test",
    srcs = [
        "log_test.go",
        "syslog_formatter_test.go",
    ],
    embed = [":go_default_library"],
    deps = [
        "//vendor/github.com/sirupsen/logrus:go_default_library",
    ],
)
