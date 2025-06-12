#!/bin/bash
# Cleanup script for the load balancer project

echo "🧹 Cleaning up project..."

# Stop all containers
echo "Stopping Docker containers..."
docker-compose down

# Remove test artifacts
echo "Removing test artifacts..."
find . -name "*.test" -delete

# Clean Java build artifacts
echo "Cleaning Java artifacts..."
(cd services/echo-java && mvn clean) || true

# Clean Node modules
echo "Cleaning Node modules..."
rm -rf services/echo-node/node_modules

echo "✨ Cleanup complete!"
