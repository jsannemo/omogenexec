load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_rust",
    strip_prefix = "rules_rust-7e7246f6c48a5d4e69744cd79b9ccb8886966ee2",
    urls = [
        # Master branch as of 2021-06-29
        "https://github.com/bazelbuild/rules_rust/archive/7e7246f6c48a5d4e69744cd79b9ccb8886966ee2.tar.gz",
    ],
)

load("@rules_rust//rust:repositories.bzl", "rust_repositories")

rust_repositories()

load("//cargo:crates.bzl", "raze_fetch_remote_crates")

raze_fetch_remote_crates()
