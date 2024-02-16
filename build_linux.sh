#!/usr/bin/env bash

OUTPATH="./bin/release"
rm -f $OUTPATH/nodelocker-linux

echo "Building..."
go build -ldflags="-w -s" -o $OUTPATH/nodelocker-linux bin/nodelocker/main.go

echo
echo "Build done, output path is: $OUTPATH/nodelocker-linux"
echo
