#!/bin/bash
# Test script for Phase 3: Observability & Operations
set -e

API_URL="http://localhost:8080"
echo "Testing Phase 3: Observability & Operations"
echo "==========================================="

# Function to check if services are ready
check_health() {
    echo "Checking health..."
    health_response=$(curl -s "$API_URL/admin/health")
    if [[ $health_response == *"\"status\":\"ok\""* ]]; then
        echo "✅ Health check passed"
        return 0
    else
        echo "❌ Health check failed"
        return 1
    fi
}

# Test 1: Metrics Collection
test_metrics() {
    echo "\nTesting metrics collection..."
    
    # Send some test requests
    for i in {1..3}; do
        curl -s -X POST "$API_URL/api" \
            -H "Content-Type: application/json" \
            -d "{\"test\":$i}" > /dev/null
    done
    
    # Check metrics
    metrics_response=$(curl -s "$API_URL/admin/metrics")
    if [[ $metrics_response == *"request_counts"* ]] && [[ $metrics_response == *"response_times"* ]]; then
        echo "✅ Metrics collection working"
        echo "Current metrics:"
        echo "$metrics_response" | python3 -m json.tool
    else
        echo "❌ Metrics collection failed"
        return 1
    fi
}

# Test 2: Backend Management
test_backend_management() {
    echo "\nTesting backend management..."
    
    # Add a new backend
    echo "Adding new backend..."
    add_response=$(curl -s -X POST "$API_URL/admin/backends" \
        -H "Content-Type: application/json" \
        -d '{"url":"http://localhost:8084"}')
    
    # Verify it was added
    backends_response=$(curl -s "$API_URL/admin/health")
    if [[ $backends_response == *"8084"* ]]; then
        echo "✅ Backend addition successful"
    else
        echo "❌ Backend addition failed"
        return 1
    fi
    
    # Remove the backend
    echo "Removing backend..."
    remove_response=$(curl -s -X DELETE "$API_URL/admin/backends" \
        -H "Content-Type: application/json" \
        -d '{"url":"http://localhost:8084"}')
    
    # Verify it was removed
    backends_response=$(curl -s "$API_URL/admin/health")
    if [[ $backends_response != *"8084"* ]]; then
        echo "✅ Backend removal successful"
    else
        echo "❌ Backend removal failed"
        return 1
    fi
}

# Main test execution
echo "Starting tests..."
if ! check_health; then
    echo "Services not ready. Make sure all containers are running:"
    echo "docker compose up -d"
    exit 1
fi

test_metrics
test_backend_management

echo "\n==========================================="
echo "✅ Phase 3 tests completed successfully"
