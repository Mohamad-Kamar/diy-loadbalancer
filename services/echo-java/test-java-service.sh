#!/bin/bash
# Simple test script for the Java Echo Service
set -e

echo "==============================================="
echo "Testing Java Echo Service (Port 8083)"
echo "==============================================="

# Function to check if service is running
check_service() {
    echo "Checking if service is running on port 8083..."
    if curl -s -f http://localhost:8083/health > /dev/null 2>&1; then
        echo "✅ Service is running"
        return 0
    else
        echo "❌ Service is not running"
        return 1
    fi
}

# Function to test health endpoint
test_health() {
    echo -n "Testing health endpoint... "
    response=$(curl -s http://localhost:8083/health)
    expected='{"status":"ok"}'
    
    if [ "$response" = "$expected" ]; then
        echo "✅ PASS"
        echo "  Response: $response"
    else
        echo "❌ FAIL"
        echo "  Expected: $expected"
        echo "  Got: $response"
        return 1
    fi
}

# Function to test echo endpoint
test_echo() {
    echo -n "Testing echo endpoint... "
    test_data='{"message":"Hello Java!","timestamp":"2024-01-01","test":true}'
    response=$(curl -s -X POST http://localhost:8083/ \
                -H 'Content-Type: application/json' \
                -d "$test_data")
    
    if [ "$response" = "$test_data" ]; then
        echo "✅ PASS"
        echo "  Input:  $test_data"
        echo "  Output: $response"
    else
        echo "❌ FAIL"
        echo "  Input:  $test_data"
        echo "  Output: $response"
        return 1
    fi
}

# Function to test method not allowed
test_method_not_allowed() {
    echo -n "Testing method not allowed... "
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X GET http://localhost:8083/)
    
    if [ "$status_code" = "405" ]; then
        echo "✅ PASS (GET / returns 405)"
    else
        echo "❌ FAIL (GET / should return 405, got $status_code)"
        return 1
    fi
}

# Main test execution
echo "Starting tests..."
echo

if check_service; then
    echo
    test_health
    echo
    test_echo
    echo
    test_method_not_allowed
    echo
    echo "==============================================="
    echo "✅ All tests passed! Java service is working correctly."
    echo "==============================================="
else
    echo
    echo "==============================================="
    echo "❌ Service is not running. Start it with:"
    echo "   cd services/echo-java"
    echo "   mvn exec:java"
    echo "or:"
    echo "   java -jar target/echo-java-0.0.1-SNAPSHOT.jar"
    echo "==============================================="
    exit 1
fi
