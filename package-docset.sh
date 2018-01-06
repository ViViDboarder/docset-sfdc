#! /bin/bash
set -e

files_dir=./resources
build_dir=./build
out_dir=.

deliverable=$1

function get-friendly-name {
    local deliverable=$1
    local name="$(tr '[:lower:]' '[:upper:]' <<< ${deliverable:0:1})${deliverable:1}"
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
    local name=$(get-friendly-name $deliverable)
    local package="$out_dir/Salesforce $name.docset"
    local icon=$(get_icon_name $deliverable)
    mkdir -p "$package/Contents/Resources/Documents"

    # Copy all meta HTML
    cp -r $build_dir/atlas.en-us.$deliverable.meta "$package/Contents/Resources/Documents/"
    # Copy HTML and CSS
    cp $build_dir/$deliverable.html "$package/Contents/Resources/Documents/"
    cp $build_dir/*.css "$package/Contents/Resources/Documents/"
    # Copy plsit
    cp $files_dir/Info-$name.plist "$package/Contents/Info.plist"
    # Copy index
    cp $build_dir/docSet.dsidx "$package/Contents/Resources/"
    # Copy icons
    cp "$files_dir/$icon.png" "$package/icon.png"
    cp "$files_dir/$icon@2x.png" "$package/icon@2x.png"

    echo "Finished building $package"
}

main
