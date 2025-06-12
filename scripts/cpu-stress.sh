#!/bin/bash
# Simulate CPU load in a running Docker container
# Usage: ./scripts/cpu-stress.sh <container_name> <percent>
# Example: ./scripts/cpu-stress.sh echo-java 80%

set -e
if [ $# -ne 2 ]; then
  echo "Usage: $0 <container_name> <percent>"
  echo "Example: $0 echo-java 80%"
  exit 1
fi
CONTAINER=$1
PERCENT=$2

# Install stress if not present
docker exec $CONTAINER sh -c "apk add --no-cache stress >/dev/null 2>&1 || apt-get update && apt-get install -y stress >/dev/null 2>&1 || true"

# Calculate number of CPUs in the container
CPUS=$(docker exec $CONTAINER nproc)
if [ -z "$CPUS" ]; then CPUS=1; fi

# Calculate number of stress workers (rounded up)
WORKERS=$(( (CPUS * ${PERCENT%%%}) / 100 ))
if [ $WORKERS -lt 1 ]; then WORKERS=1; fi

echo "Stressing $CONTAINER with $WORKERS CPU workers ($PERCENT of $CPUS CPUs)..."
docker exec -d $CONTAINER stress --cpu $WORKERS --timeout 300

echo "CPU stress started for 5 minutes. To stop early:"
echo "  docker exec $CONTAINER pkill stress || true"
