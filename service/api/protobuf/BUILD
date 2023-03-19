load("@rules_proto//proto:defs.bzl", "proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

proto_library(
    name = "messages_proto",
    srcs = ["messages.proto"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_google_protobuf//:struct_proto",
        "@com_google_protobuf//:timestamp_proto",
    ],
)

go_proto_library(
    name = "messages_go_proto",
    importpath = "github.com/aeekayy/stilla/service/api/protobuf/messages",
    proto = ":messages_proto",
    visibility = ["//visibility:public"],
)

go_library(
    name = "messages",
    embed = [":messages_go_proto"],
    importpath = "github.com/aeekayy/stilla/service/api/protobuf/messages",
    visibility = ["//visibility:public"],
)
