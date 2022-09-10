#!/usr/bin/env bash

base_path=`dirname -- "$0"`/..

branch=$(git symbolic-ref HEAD | sed -e 's,.*/\(.*\),\1,')
if [ $branch != "master" ]
then
        echo "Releases can only be made from master"
        exit 1
fi

require_clean_work_tree () {
    # Update the index
    git update-index -q --ignore-submodules --refresh
    err=0

    # Disallow unstaged changes in the working tree
    if ! git diff-files --quiet --ignore-submodules --
    then
        echo >&2 "cannot $1: you have unstaged changes."
        git diff-files --name-status -r --ignore-submodules -- >&2
        err=1
    fi

    # Disallow uncommitted changes in the index
    if ! git diff-index --cached --quiet HEAD --ignore-submodules --
    then
        echo >&2 "cannot $1: your index contains uncommitted changes."
        git diff-index --cached --name-status -r --ignore-submodules HEAD -- >&2
        err=1
    fi

    if [ $err = 1 ]
    then
        echo >&2 "Please commit or stash them."
        exit 1
    fi
}


require_clean_work_tree "deploy"

git fetch origin 
ahead=`git rev-list --left-right --count master...origin/master | cut -f1`
behind=`git rev-list --left-right --count master...origin/master | cut -f2`

if [ $ahead != "0" ]
then
    echo "You are $ahead commits ahead of master: please push your change"
    exit 1
fi

$base_path/packaging/build-omogenexec.sh
rm -f $base_path/packaging/omogenexec-debian.deb
cp $base_path/bazel-bin/deb/sandbox/omogenexec-debian.deb $base_path/packaging

gh auth status || gh auth login

echo "Last release $(git tag | sort | tail -n 1)"
read -p "New tag: " tag
gh release create $tag 'packaging/omogenexec-debian.deb' --notes ""
