#!/usr/bin/env bash
# deploy.sh — build and run the server, auto-incrementing the patch version.
set -euo pipefail

VERSION_FILE="VERSION"

# Read current version
VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Bump patch
PATCH=$((PATCH + 1))
NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "$NEW_VERSION" > "$VERSION_FILE"
echo "Version: $NEW_VERSION"

# Kill existing server on :8080 if running
EXISTING_PID=$(lsof -ti :8080 2>/dev/null || true)
if [ -n "$EXISTING_PID" ]; then
    echo "Stopping existing server (PID $EXISTING_PID)..."
    kill $EXISTING_PID 2>/dev/null || true
    sleep 1
fi

# Build with version baked in
echo "Building server..."
go build -ldflags "-X main.Version=${NEW_VERSION}" -o server ./cmd/server/

# Run smoke tests
echo "Running smoke tests..."
go test ./pkg/server/ -run TestSmoke -v -timeout 120s || exit 1

# Run
echo "Starting server on :8080..."
./server "$@"
