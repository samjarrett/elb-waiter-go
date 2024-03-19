#!/bin/bash -eu
set -o pipefail

OSES="linux darwin"
ARCHS="amd64 arm64"

rm -f build/*
for GOOS in $OSES; do
    for GOARCH in $ARCHS; do
        echo ">> Building ${GOOS}/${GOARCH}"
        export GOOS
        export GOARCH
        time go build -o "build/elb-waiter_${GOOS}_${GOARCH}"
        echo ""
    done
done
