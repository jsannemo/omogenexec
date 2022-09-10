#!/usr/bin/env bash

base_path=`dirname -- "$0"`/..
(cd $base_path; bazel build ...;)
rm -f $base_path/packaging/omogenexec-debian.deb
cp $base_path/bazel-bin/deb/sandbox/omogenexec-debian.deb $base_path/packaging
