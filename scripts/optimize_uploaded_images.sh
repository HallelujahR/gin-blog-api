#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")/.."

GOCACHE="${GOCACHE:-/tmp/blog-api-gocache}" \
go run ./cmd/tools/optimize_uploaded_images "$@"
