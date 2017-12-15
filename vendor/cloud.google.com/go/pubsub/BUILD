load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "flow_controller.go",
        "iterator.go",
        "message.go",
        "pubsub.go",
        "service.go",
        "snapshot.go",
        "subscription.go",
        "topic.go",
    ],
    importpath = "cloud.google.com/go/pubsub",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/iam:go_default_library",
        "//vendor/cloud.google.com/go/internal/version:go_default_library",
        "//vendor/cloud.google.com/go/pubsub/apiv1:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/sync/errgroup:go_default_library",
        "//vendor/golang.org/x/sync/semaphore:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/support/bundler:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/pubsub/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/status:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "endtoend_test.go",
        "fake_test.go",
        "flow_controller_test.go",
        "integration_test.go",
        "service_test.go",
        "streaming_pull_test.go",
        "subscription_test.go",
        "topic_test.go",
        "utils_test.go",
    ],
    importpath = "cloud.google.com/go/pubsub",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/iam:go_default_library",
        "//vendor/cloud.google.com/go/internal:go_default_library",
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/sync/errgroup:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/pubsub/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_subscription_iterator_test.go",
        "example_test.go",
        "example_topic_iterator_test.go",
    ],
    importpath = "cloud.google.com/go/pubsub_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
    ],
)
