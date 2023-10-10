#!/usr/bin/env bash

rm -rf builds
mkdir builds

ARCHS=('arm' 'arm64' '386' 'amd64')
for ARCH in "${ARCHS[@]}"
do
    printf "compiling %s\n" "${ARCH}"
    CGO_ENABLED=0 GOOS="linux" GOARCH="${ARCH}" go build -trimpath=true -ldflags="-s -w" -o builds/geteduroam-cli-linux-"${ARCH}" ./cmd/geteduroam-cli
done

