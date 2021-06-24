#!/bin/sh

echo "Configuring test env..."
apk add --no-cache gcc musl-dev

if test "$1" == "test"; then
	echo "testing fixtures..."
    go test -tags musl,static,netgo -mod=vendor ./...
elif test "$1" == "vet"; then
	echo "vetting build..."
	go vet -tags musl,static,netgo -mod=vendor ./*.go
fi
