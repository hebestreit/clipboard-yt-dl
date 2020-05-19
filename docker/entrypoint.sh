#!/usr/bin/env bash

set -xe

PROJECT_ROOT="${1}"
PACKAGE="${2}"
PACKAGE_BASENAME="$(basename "${2}")"
BUILD_OUTPUT="/go/src/${PROJECT_ROOT}/${3}/${PACKAGE_BASENAME}"

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
