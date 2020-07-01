#! /bin/bash

set -eu
set -o pipefail

if ! git config --global user.name ;then
    # If no global git configs exist, let's set some temporary values
    export GIT_COMMITTER_NAME="${GIT_COMMITTER_NAME:-"ViViDboarder"}"
    export GIT_COMMITTER_EMAIL="${GIT_COMMITTER_EMAIL:-"ViViDboarder@gmail.com"}"
    export GIT_AUTHOR_NAME="$GIT_COMMITTER_NAME"
    export GIT_AUTHOR_EMAIL="$GIT_COMMITTER_EMAIL"
fi
# Get name of the fork and target repo
FORK_REPO="${FORK_REPO:-"ViViDboarder/Dash-User-Contributions"}"
TARGET_REPO="${TARGET_REPO:-"ViViDboarder/Dash-User-Contributions"}"

# If no github user is provided, take it from the fork name
if [ -z "${GITHUB_USER:-""}" ]; then
    GITHUB_USER="${FORK_REPO%%/*}"
fi
GITHUB_TOKEN="${GITHUB_TOKEN:-default}"

WORKDIR=$(pwd)
TMP_DIR="$WORKDIR/repotmp"
REPO_DIR="$TMP_DIR/Dash-User-Contributions"

function validate() {
    if [ -z "$GITHUB_TOKEN" ]; then
        echo "Must provide \$GITHUB_TOKEN as an environment variable"
        exit 1
    fi

    echo "Creating PR for $GITHUB_USER to $TARGET_REPO"
}

function read_version() {
    local apex_version
    apex_version="$(cat ./build/apexcode-version.txt)"
    local pages_version
    pages_version="$(cat ./build/pages-version.txt)"
    local lightning_version
    lightning_version="$(cat ./build/lightning-version.txt)"

    if [ "$apex_version" != "$pages_version" ] || [ "$apex_version" != "$lightning_version" ]; then
        echo "Apex: $apex_version, Pages: $pages_version, Lightning: $lightning_version"
        echo "One of the doc versions doesn't match"
        exit 1
    fi
    # All versions match, return one of them
    echo "$apex_version"
}

function workdir_git() {
    cd "$REPO_DIR"
    if ! git "$@" ;then
        # Be sure to return to workdir after a failed git command
        cd "$WORKDIR"
        return 1
    fi
    cd "$WORKDIR"
}

function shallow_clone_or_pull() {
    mkdir -p "$TMP_DIR"
    if [ -d "$REPO_DIR" ]; then
        workdir_git checkout master
        workdir_git pull --ff-only origin master
    else
        git clone --depth 1 "https://$GITHUB_USER:$GITHUB_TOKEN@github.com/$FORK_REPO" "$REPO_DIR"
    fi
}

function copy_release() {
    cp -r archive/* "$REPO_DIR/docsets/"
}

function create_release_branch() {
    local branch="$1"
    if ! workdir_git checkout -b "$branch" ; then
        echo "Could not create release branch. Release likely already exists."
        exit 1
    fi
}

function create_pr() {
    local version="$1"
    local branch="$2"
    local title="Update Salesforce docsets to $version"
    workdir_git checkout "$branch"
    workdir_git add .
    workdir_git commit -m "$title"
    workdir_git push origin HEAD
    local result
    result=$(curl \
        -u "$GITHUB_USER:$GITHUB_TOKEN" \
        -X POST \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/$TARGET_REPO/pulls" \
        -d "{\"title\":\"$title\",\"body\":\"This branch contains auto-generated updates to version $version\", \"head\":\"$GITHUB_USER:$branch\",\"base\":\"master\"}")


    local result_url
    if result_url="$(echo "$result" | jq --exit-status --raw-output .html_url)" ;then
        echo "Pull request created at $result_url"
    else
        echo "$result"
        exit 1
    fi
}

function main() {
    validate

    local version
    version="$(read_version)" || { echo "$version"; exit 1; }
    local branch="salesforce-$version"

    shallow_clone_or_pull
    copy_release
    create_release_branch "$branch"
    create_pr "$version" "$branch"
}

main
