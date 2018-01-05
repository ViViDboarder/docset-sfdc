#! /bin/bash
set -e

files_dir=./SFDashC
build_dir=./build

name=$1
deliverable=$(echo $name | tr '[:upper:]' '[:lower:]')
if [ "$deliverable" == "apex" ]; then
    deliverable="apexcode"
fi

package="Salesforce $name.docset"
version=$(cat $build_dir/$deliverable-version.txt)

cat $files_dir/docset-$deliverable.json | sed s/VERSION/$version/ > $build_dir/docset-$deliverable.json
mkdir -p "$package/Contents/Resources/Documents"
cp -r $build_dir/atlas.en-us.$deliverable.meta "$package/Contents/Resources/Documents/"
cp $build_dir/*.html "$package/Contents/Resources/Documents/"
cp $build_dir/*.css "$package/Contents/Resources/Documents/"
cp $files_dir/Info-$name.plist "$package/Contents/Info.plist"
cp $build_dir/docSet.dsidx "$package/Contents/Resources/"

echo "Finished building $package"
