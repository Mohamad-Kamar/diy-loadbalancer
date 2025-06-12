#!/bin/bash
# Main test runner script for the load balancer project

set -e  # Exit on any error

echo "🚀 Starting test suite..."

# Function to print section headers
print_header() {
    echo "================================================"
    echo "🔍 $1"
    echo "================================================"
}

# Start all services
print_header "Starting services"
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 5

# Run unit tests for each service
print_header "Running unit tests"

echo "Testing Round Robin API (Go)..."
(cd services/round-robin-api && go test ./... -v) || exit 1

echo "Testing Echo Go Service..."
(cd services/echo-go && go test ./... -v) || exit 1

echo "Testing Echo Node Service..."
(cd services/echo-node && npm test) || exit 1

echo "Testing Echo Java Service..."
(cd services/echo-java && mvn test) || exit 1

# Run integration tests
print_header "Running integration tests"
./scripts/run-all-tests.sh

# Run load tests if hey is installed
if command -v hey &> /dev/null; then
    print_header "Running load tests"
    hey -n 1000 -c 10 -m POST \
        -H "Content-Type: application/json" \
        -d '{"test":"load"}' \
        http://localhost:8080/api
else
    echo "⚠️  hey is not installed, skipping load tests"
    echo "To install hey: go install github.com/rakyll/hey@latest"
fi

# Clean up
print_header "Cleanup"
echo "Stopping services..."
docker-compose down

echo "✅ All tests completed!"
