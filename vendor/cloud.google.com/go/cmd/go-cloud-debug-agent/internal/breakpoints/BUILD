load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["breakpoints.go"],
    importpath = "cloud.google.com/go/cmd/go-cloud-debug-agent/internal/breakpoints",
    visibility = ["//vendor/cloud.google.com/go/cmd/go-cloud-debug-agent:__subpackages__"],
    deps = [
        "//vendor/golang.org/x/debug:go_default_library",
        "//vendor/google.golang.org/api/clouddebugger/v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["breakpoints_test.go"],
    importpath = "cloud.google.com/go/cmd/go-cloud-debug-agent/internal/breakpoints",
    library = ":go_default_library",
    deps = [
        "//vendor/golang.org/x/debug:go_default_library",
        "//vendor/google.golang.org/api/clouddebugger/v2:go_default_library",
    ],
)
