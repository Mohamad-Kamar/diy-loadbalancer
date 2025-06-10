#!/bin/bash
# Script to run the round-robin load balancer with environment variables

# Change to the script's directory
cd "$(dirname "$0")"

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "📝 Please copy .env.example to .env and configure your backends:"
    echo "   cp .env.example .env"
    echo "   # Edit .env with your backend URLs"
    exit 1
fi

# Load environment variables from .env file
echo "📝 Loading environment variables from .env..."
export $(grep -v '^#' .env | xargs)

# Validate BACKENDS is set
if [ -z "$BACKENDS" ]; then
    echo "❌ BACKENDS environment variable is not set in .env file"
    exit 1
fi

echo "🚀 Starting Round Robin Load Balancer..."
echo "📍 Backends: $BACKENDS"
echo "🔗 Load balancer will be available at: http://localhost:8080"
echo "---"

# Run the Go application
go run cmd/main.go
