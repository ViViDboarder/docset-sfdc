#! /bin/bash
set -e

files_dir=./resources
build_dir=./build
out_dir=.
archive_dir=./archive

deliverable=$1

function get_friendly_name {
    local deliverable=$1
    local name="$(tr '[:lower:]' '[:upper:]' <<< "${deliverable:0:1}")${deliverable:1}"
    case "$deliverable" in
        "apexcode")
            name="Apex"
            ;;
        "pages")
            name="Visualforce"
            ;;
    esac

    echo $name
}

function get_icon_name {
    local icon="cloud-icon"
    case "$1" in
        "lightning")
            icon="bolt-icon"
            ;;
    esac
    echo $icon
}

function main {
    local name=$(get_friendly_name "$deliverable")
    local package="$out_dir/Salesforce $name.docset"
    local archive_dir="$archive_dir/Salesforce_$name"
    local archive="$archive_dir/Salesforce_$name.tgz"
    local icon=$(get_icon_name "$deliverable")
    mkdir -p "$archive_dir"

    # Generate docset.json
    version=$(cat "$build_dir/$deliverable-version.txt")
    sed "s/VERSION/$version/" "$files_dir/docset-$deliverable.json" > "$archive_dir/docset.json"
    # Generated tgz archive
    tar --exclude=".DS_Store" -czf "$archive" "$package"
    # Copy icons
    cp "$files_dir/$icon.png" "$archive_dir/icon.png"
    cp "$files_dir/$icon@2x.png" "$archive_dir/icon@2x.png"
    # Copy readme
    sed "s/DOCSET_NAME/$name/" "$files_dir/Archive_Readme.md" > "$archive_dir/README.md"

    echo "Finished archive $archive"
}

main
