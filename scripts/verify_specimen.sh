#!/usr/bin/env sh
set -eu

echo "==> Running go vet on refactor/..."
go vet ./refactor/...

echo "==> Running race-enabled unit tests on refactor/..."
CGO_ENABLED=1 go test -race -v ./refactor/...

echo "==> Verifying static CGO_ENABLED=0 build..."
CGO_ENABLED=0 go build ./refactor/...

echo "==> All specimen invariants verified!"
