load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

package(
    default_visibility = ["//visibility:public"],
)

go_binary(
    name = "stilla",
    embed = [":service_lib"],
)

go_library(
    name = "service_lib",
    srcs = ["main.go"],
    importpath = "github.com/aeekayy/stilla/service",
    deps = [
        "//service/pkg/api",
        "//service/pkg/models",
        "@com_github_getsentry_sentry_go//:go_default_library",
        "@org_uber_go_ratelimit//:go_default_library",
        "@org_uber_go_zap//:go_default_library",
    ],
)
