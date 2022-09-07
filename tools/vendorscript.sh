#!/bin/sh

DEFAULT_TAGS=static,netgo

if which apk > /dev/null 2>&1; then
	apk add --no-cache gcc musl-dev
fi
rm -rf /app/vendor /app/go.mod /app/go.sum
go mod init "${PACKAGE_NAME}"
go mod tidy
go mod vendor

