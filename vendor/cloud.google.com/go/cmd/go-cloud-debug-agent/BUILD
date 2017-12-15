load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = select({
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "debuglet.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "cloud.google.com/go/cmd/go-cloud-debug-agent",
    visibility = ["//visibility:private"],
    deps = select({
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "//vendor/cloud.google.com/go/cmd/go-cloud-debug-agent/internal/breakpoints:go_default_library",
            "//vendor/cloud.google.com/go/cmd/go-cloud-debug-agent/internal/controller:go_default_library",
            "//vendor/cloud.google.com/go/cmd/go-cloud-debug-agent/internal/valuecollector:go_default_library",
            "//vendor/cloud.google.com/go/compute/metadata:go_default_library",
            "//vendor/golang.org/x/debug:go_default_library",
            "//vendor/golang.org/x/debug/local:go_default_library",
            "//vendor/golang.org/x/net/context:go_default_library",
            "//vendor/golang.org/x/oauth2:go_default_library",
            "//vendor/golang.org/x/oauth2/google:go_default_library",
            "//vendor/google.golang.org/api/clouddebugger/v2:go_default_library",
        ],
        "//conditions:default": [],
    }),
)

go_binary(
    name = "go-cloud-debug-agent",
    importpath = "cloud.google.com/go/cmd/go-cloud-debug-agent",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
