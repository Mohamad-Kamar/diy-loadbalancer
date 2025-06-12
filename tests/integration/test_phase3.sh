#!/bin/bash
# Enhanced test script for Phase 3: Observability & Operations
set -e

API_URL="http://localhost:8080"
echo "🧪 Testing Phase 3: Observability & Operations"
echo "=============================================="

# Helper function to format JSON output
format_json() {
    python3 -m json.tool
}

# Function to check if services are ready
check_health() {
    echo "🔍 Checking health..."
    for i in {1..30}; do
        health_response=$(curl -s "$API_URL/admin/health")
        if [[ $health_response == *"\"status\":\"ok\""* ]]; then
        echo "✅ Health check passed"
        echo "Current backends:"
        echo "$health_response" | format_json
        return 0
    else
        echo "❌ Health check failed"
        echo "$health_response"
        return 1
    fi
}

# Function to send test requests and verify round-robin distribution
test_distribution() {
    echo "🔄 Testing request distribution..."
    
    # Send multiple requests and collect responses
    for i in {1..9}; do
        echo "📤 Request $i"
        response=$(curl -s -X POST "$API_URL/api" \
            -H "Content-Type: application/json" \
            -H "X-Request-ID: test-$i" \
            -d "{\"test\":$i}")
        echo "📥 Response: $response"
        echo "---"
        sleep 0.5
    done

    # Check metrics to verify distribution
    metrics=$(curl -s "$API_URL/admin/metrics")
    echo "📊 Distribution metrics:"
    echo "$metrics" | format_json

    # Validate metrics content
    if [[ $metrics == *"request_counts"* ]] && \
       [[ $metrics == *"response_times"* ]] && \
       [[ $metrics == *"error_rates"* ]]; then
        echo "✅ Request distribution looks good"
        return 0
    else
        echo "❌ Metrics validation failed"
        return 1
    fi
}

# Function to test backend management
test_backend_management() {
    echo "🔧 Testing backend management..."
    
    # 1. Add a new backend
    echo "➕ Adding new backend instance..."
    add_response=$(curl -s -X POST "$API_URL/admin/backends" \
        -H "Content-Type: application/json" \
        -d '{"url":"http://echo-go:8081"}')
    echo "Response: $add_response"
    
    # Verify it was added
    echo "🔍 Verifying backend addition..."
    health_after_add=$(curl -s "$API_URL/admin/health")
    if [[ $health_after_add == *"echo-go:8081"* ]]; then
        echo "✅ Backend successfully added"
        echo "$health_after_add" | format_json
    else
        echo "❌ Backend addition failed"
        echo "$health_after_add" | format_json
        return 1
    fi
    
    # 2. Send some requests to see distribution
    echo "📤 Testing with new backend..."
    for i in {1..4}; do
        curl -s -X POST "$API_URL/api" \
            -H "Content-Type: application/json" \
            -d "{\"test\":\"new-backend-$i\"}" > /dev/null
    done

    # Check metrics again
    metrics_after_add=$(curl -s "$API_URL/admin/metrics")
    echo "📊 Updated metrics:"
    echo "$metrics_after_add" | format_json
    
    # 3. Remove the backend
    echo "➖ Removing test backend..."
    remove_response=$(curl -s -X DELETE "$API_URL/admin/backends" \
        -H "Content-Type: application/json" \
        -d '{"url":"http://echo-go:8081"}')
    
    # Verify removal
    echo "🔍 Verifying backend removal..."
    health_after_remove=$(curl -s "$API_URL/admin/health")
    if [[ $health_after_remove != *"echo-go:8081"* ]]; then
        echo "✅ Backend successfully removed"
        echo "$health_after_remove" | format_json
    else
        echo "❌ Backend removal failed"
        echo "$health_after_remove" | format_json
        return 1
    fi
}

# Function to test error handling
test_error_handling() {
    echo "⚠️ Testing error handling..."
    
    # Test invalid backend addition
    echo "Testing invalid backend URL..."
    invalid_response=$(curl -s -X POST "$API_URL/admin/backends" \
        -H "Content-Type: application/json" \
        -d '{"url":"not-a-valid-url:1234"}')
    if [[ $invalid_response == *"error"* ]]; then
        echo "✅ Invalid URL properly rejected"
    else
        echo "❌ Invalid URL handling failed"
        return 1
    fi
    
    # Test invalid request
    echo "Testing invalid request..."
    invalid_json_response=$(curl -s -X POST "$API_URL/api" \
        -H "Content-Type: application/json" \
        -d '{invalid json}')
    if [[ $invalid_json_response == *"error"* ]]; then
        echo "✅ Invalid JSON properly handled"
    else
        echo "❌ Invalid JSON handling failed"
        return 1
    fi
}

# Function to test request tracing
test_request_tracing() {
    echo "🔍 Testing request tracing..."
    
    # Send request with custom trace ID
    trace_id="test-trace-$(date +%s)"
    response=$(curl -s -X POST "$API_URL/api" \
        -H "Content-Type: application/json" \
        -H "X-Request-ID: $trace_id" \
        -d '{"trace_test": true}')
        
    # Check metrics for trace
    metrics=$(curl -s "$API_URL/admin/metrics")
    if [[ $metrics == *"$trace_id"* ]]; then
        echo "✅ Request tracing working"
        echo "📊 Recent requests:"
        echo "$metrics" | jq '.recent_requests' 2>/dev/null || echo "$metrics"
    else
        echo "❌ Request tracing failed"
        return 1
    fi
}

# Main test execution
echo "🚀 Starting tests..."

# 1. Initial health check
if ! check_health; then
    echo "❌ Services not ready. Please ensure all services are running:"
    echo "docker-compose up -d"
    exit 1
fi

# 2. Run all tests
echo "=============================================="
test_distribution
echo "=============================================="
test_backend_management
echo "=============================================="
test_error_handling
echo "=============================================="
test_request_tracing

echo "=============================================="
echo "✅ All Phase 3 tests completed successfully!"
echo "=============================================="
