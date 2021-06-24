#!/bin/sh

echo "selecting build environment..."
if which apk 2>&1; then
	apk add --no-cache gcc musl-dev
	go test -tags musl ./...
	go vet -tags musl ./*.go
	go build -tags musl,static,netgo -o "$1-musl" ./...
else
	go test ./...
	go vet ./*.go
	go build -o "$1-libc" -tags static,netgo  ./...
fi
