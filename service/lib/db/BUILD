load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "lib",
    srcs = ["db.go"],
    visibility = ["//visibility:public"],
)

go_library(
    name = "db",
    srcs = ["db.go"],
    importpath = "github.com/aeekayy/stilla/service/lib/db",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_google_uuid//:go_default_library",
        "@com_github_jackc_pgx_v5//:go_default_library",
        "@com_github_jackc_pgx_v5//pgconn:go_default_library",
        "@com_github_jackc_pgx_v5//pgxpool:go_default_library",
        "@org_golang_x_exp//slices:go_default_library",
        "@org_mongodb_go_mongo_driver//mongo:go_default_library",
        "@org_mongodb_go_mongo_driver//mongo/options:go_default_library",
    ],
)

go_test(
    name = "db_test",
    srcs = ["db_test.go"],
    embed = [":db"],
    deps = [
        "//service/pkg/utils:lib",
        "@com_github_stretchr_testify//assert:go_default_library",
    ],
)
