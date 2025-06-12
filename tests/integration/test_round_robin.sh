#!/bin/bash
# Simple integration test for round-robin and echo
set -e

API_URL="http://localhost:8080/api"

echo "Testing Round Robin Load Balancer..."
echo "Sending 6 POST requests to $API_URL..."
echo "============================================"

for i in {1..6}; do
  echo "Request $i:"
  response=$(curl -s -X POST "$API_URL" -H 'Content-Type: application/json' -d '{"msg": "hello", "request_id":'$i'}')
  echo "Response: $response"
  echo "---"
  sleep 0.5
done

echo "============================================"
echo "Test completed. You should see responses from all three backends (Go:8081, Node:8082, Java:8083)"
