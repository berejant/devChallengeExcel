#!/usr/bin/env sh

set -e

echo -n "Download modules..."
go mod download
echo "Done."

echo "Run unit tests..."
go test . -tags test -covermode=atomic -coverprofile /tmp/coverage.out

go tool cover -func=/tmp/coverage.out

echo "Done."
