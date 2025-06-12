#!/bin/bash
# Run all integration tests for diy-loadbalancer
set -e

echo "============================================="
echo "Running all integration tests..."
echo "============================================="

for test in tests/integration/*.sh; do
  echo "\n--- Running $test ---"
  bash "$test"
done

echo "============================================="
echo "All tests completed!"
echo "============================================="
