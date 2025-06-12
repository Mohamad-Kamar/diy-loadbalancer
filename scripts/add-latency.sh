#!/bin/bash
# Add network latency to a running Docker container (simulate slow backend)
# Usage: ./scripts/add-latency.sh <container_name> <latency>
# Example: ./scripts/add-latency.sh echo-node 500ms

set -e
if [ $# -ne 2 ]; then
  echo "Usage: $0 <container_name> <latency>"
  echo "Example: $0 echo-node 500ms"
  exit 1
fi
CONTAINER=$1
LATENCY=$2

echo "Adding $LATENCY latency to container $CONTAINER..."
docker exec $CONTAINER sh -c "apk add --no-cache iproute2 >/dev/null 2>&1 || true"
docker exec $CONTAINER tc qdisc add dev eth0 root netem delay $LATENCY || {
  echo "Failed to add latency. It may already be set."
}
echo "Done. To remove latency:"
echo "  docker exec $CONTAINER tc qdisc del dev eth0 root netem || true"
