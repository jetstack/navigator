load("@io_bazel_rules_go//go:def.bzl", "gazelle")

gazelle(
    name = "gazelle",
    args = ["-build_file_name", "BUILD,BUILD.bazel", "-proto", "legacy"],
    prefix = "github.com/jetstack/navigator",
    external = "vendored",
)
