#!/bin/sh

DEFAULT_TAGS=static,netgo

echo "selecting build environment..."
if which apk > /dev/null 2>&1; then
	apk add --no-cache gcc musl-dev
	go build -tags "musl,${DEFAULT_TAGS}" -ldflags='-w -s -extldflags "-static"' -o "$1-musl" ./...
else
	go build -o "$1-libc" -tags "${DEFAULT_TAGS}" -ldflags='-w -s -extldflags "-static"' ./...
fi
