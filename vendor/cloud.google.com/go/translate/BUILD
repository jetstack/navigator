load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["translate.go"],
    importpath = "cloud.google.com/go/translate",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/internal/version:go_default_library",
        "//vendor/cloud.google.com/go/translate/internal/translate/v2:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/text/language:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport/http:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["translate_test.go"],
    importpath = "cloud.google.com/go/translate",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/text/language:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["examples_test.go"],
    importpath = "cloud.google.com/go/translate_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/text/language:go_default_library",
    ],
)
