load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "models",
    srcs = ["config.go"],
    importpath = "github.com/aeekayy/stilla/service/pkg/models",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_confluentinc_confluent_kafka_go//kafka:go_default_library",
        "@com_github_spf13_viper//:go_default_library",
    ],
)
