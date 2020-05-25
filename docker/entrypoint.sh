#!/usr/bin/env bash

set -xe

PACKAGE="${1}"
BUILD_OUTPUT="/go/src/${PACKAGE}/${2}"

shift
shift
shift

cd "${PROJECT_ROOT}"

export CGO_ENABLED=1
export GOARCH=amd64

echo "Building linux binary"
GOOS=linux CC=clang CXX=clang++ go build -o "${BUILD_OUTPUT}" ${PACKAGE} $*

echo "Building windows binary"
GOOS=windows CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build -o "${BUILD_OUTPUT}.exe" -ldflags "-H=windowsgui -extldflags=-s" ${PACKAGE} $*
