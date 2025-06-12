#!/bin/bash
# Quick test script for development - runs only tests in the changed service

set -e

# Determine which service has changed based on git status
changed_files=$(git diff --name-only HEAD)

run_go_tests() {
    local service=$1
    echo "🔍 Running Go tests for $service..."
    (cd "services/$service" && go test ./... -v)
}

run_node_tests() {
    echo "🔍 Running Node.js tests..."
    (cd services/echo-node && npm test)
}

run_java_tests() {
    echo "🔍 Running Java tests..."
    (cd services/echo-java && mvn test)
}

# Check each service
if echo "$changed_files" | grep -q "services/round-robin-api/"; then
    run_go_tests "round-robin-api"
fi

if echo "$changed_files" | grep -q "services/echo-go/"; then
    run_go_tests "echo-go"
fi

if echo "$changed_files" | grep -q "services/echo-node/"; then
    run_node_tests
fi

if echo "$changed_files" | grep -q "services/echo-java/"; then
    run_java_tests
fi

# If no service-specific changes, run integration tests
if ! echo "$changed_files" | grep -q "services/"; then
    echo "🔍 Running integration tests..."
    ./scripts/run-all-tests.sh
fi

echo "✅ Quick tests completed!"
