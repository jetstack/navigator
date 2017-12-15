load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["client.go"],
    importpath = "cloud.google.com/go/cmd/go-cloud-debug-agent/internal/controller",
    visibility = ["//vendor/cloud.google.com/go/cmd/go-cloud-debug-agent:__subpackages__"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/google.golang.org/api/clouddebugger/v2:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport/http:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["client_test.go"],
    importpath = "cloud.google.com/go/cmd/go-cloud-debug-agent/internal/controller",
    library = ":go_default_library",
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/google.golang.org/api/clouddebugger/v2:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
    ],
)
