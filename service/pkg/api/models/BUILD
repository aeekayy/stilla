load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "models",
    srcs = [
        "model_audit_log.go",
        "model_config_in.go",
        "model_config_response.go",
        "model_config_store.go",
        "model_config_version.go",
        "model_error.go",
        "model_healthcheck.go",
        "model_host_login_in.go",
        "model_host_register_in.go",
        "model_id_response.go",
        "model_update_config_in.go",
    ],
    importpath = "github.com/aeekayy/stilla/service/pkg/api/models",
    visibility = ["//visibility:public"],
    deps = [
        "@org_mongodb_go_mongo_driver//bson",
        "@org_mongodb_go_mongo_driver//bson/primitive",
    ],
)

go_test(
    name = "models_test",
    srcs = ["model_config_response_test.go"],
    embed = [":models"],
    deps = [
        "@com_github_stretchr_testify//assert",
        "@org_mongodb_go_mongo_driver//bson",
    ],
)
