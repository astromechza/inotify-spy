#!/usr/bin/env bash

set -e

# main func
function buildbinary {
    goos=$1
    goarch=$2

    echo "Building official $goos $goarch binary"

    outputfolder="build/${goos}_${goarch}"
    echo "Output Folder $outputfolder"
    mkdir -pv $outputfolder

    export GOOS=$goos
    export GOARCH=$goarch

    go build -i -v -o "$outputfolder/inotify-spy" github.com/AstromechZA/inotify-spy

    echo "Done"
    ls -l "$outputfolder/inotify-spy"
    file "$outputfolder/inotify-spy"
    echo
}

# build for mac 64bit
buildbinary darwin amd64

# build for linux 64bit
buildbinary linux amd64

# build for linux 32 bit
buildbinary linux 386

# see https://golang.org/doc/install/source#environment for more goos/goarch variables
