# diy-loadbalancer

## Usage

### Option 1: Docker (Recommended)

```bash
docker-compose up --build
```

### Option 2: Manual Setup (Development)

1. **Start the backend echo services** (in separate terminals):

```bash
# Terminal 1: Go echo service
cd services/echo-go
go run main.go

# Terminal 2: Node.js echo service  
cd services/echo-node
node index.js

# Terminal 3: Java echo service
cd services/echo-java
mvn exec:java
```

2. **Start the round-robin load balancer**:

```bash
# Terminal 4: Load balancer
cd services/round-robin-api
./run.sh
```

Or manually with environment variables:
```bash
cd services/round-robin-api
BACKENDS=http://localhost:8081,http://localhost:8082,http://localhost:8083 go run cmd/main.go
```

## Managing Backends

The load balancer supports dynamic backend management through its Admin API. This allows you to add or remove backend services without restarting the load balancer.

### Adding a New Backend

There are several ways to add a new backend service:

#### 1. Using an Existing Service Type

You can start additional instances of any existing service (Go, Node.js, or Java):

```bash
# Example: Starting a new Go backend

# 1. Start the container with a unique name and port
docker run -d -p 8085:8081 --name echo-go-new diy-loadbalancer-echo-go

# 2. Connect it to the load balancer network
docker network connect diy-loadbalancer_default echo-go-new

# 3. Add it to the load balancer using the container name
curl -X POST http://localhost:8080/admin/backends \
  -H "Content-Type: application/json" \
  -d '{"url":"http://echo-go-new:8085"}'
```

#### 2. Custom Backend Service

You can also add your own backend service as long as it implements:
- A POST `/` endpoint that echoes back JSON payloads
- A GET `/health` endpoint that returns `{"status":"ok"}`

Example steps for a custom service:
```bash
# 1. Build and start your service
docker build -t my-custom-backend ./my-service
docker run -d -p 8090:8080 --name my-backend my-custom-backend

# 2. Connect to the load balancer network
docker network connect diy-loadbalancer_default my-backend

# 3. Add to the load balancer (use container name)
curl -X POST http://localhost:8080/admin/backends \
  -H "Content-Type: application/json" \
  -d '{"url":"http://my-backend:8080"}'
```

### URL Format Requirements

When adding backends, follow these URL format guidelines:

1. **Docker Networking**:
   - Use container names as hostnames (e.g., `http://echo-go-new:8081`)
   - Avoid using `localhost` as it won't work inside containers
   - Make sure the container is on the same Docker network

2. **Direct Networking**:
   - Use actual hostnames or IPs when not using Docker
   - Include the port (e.g., `http://localhost:8085`)
   - Must be reachable from the load balancer container

### Verifying Backend Status

After adding a backend, verify it's working:

```bash
# 1. Check if backend is listed
curl http://localhost:8080/admin/health

# 2. Send test requests
curl -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"test":"payload"}'

# 3. Check metrics to verify load distribution
curl http://localhost:8080/admin/metrics
```

### Removing a Backend

To remove a backend service:

```bash
# 1. Remove from load balancer (same URL used when adding)
curl -X DELETE http://localhost:8080/admin/backends \
  -H "Content-Type: application/json" \
  -d '{"url":"http://echo-go-new:8081"}'

# 2. Stop and remove the container (if using Docker)
docker rm -f echo-go-new

# 3. Verify removal
curl http://localhost:8080/admin/health
```

### Best Practices

1. **Health Checks**:
   - Backends are automatically health checked every 5 seconds
   - Unhealthy backends are excluded from load balancing
   - Health endpoint should respond quickly (<1s)

2. **Gradual Changes**:
   - Add/remove one backend at a time
   - Verify health status between changes
   - Monitor metrics for proper request distribution

3. **Resource Management**:
   - Consider container resource limits (CPU/memory)
   - Monitor backend performance via metrics
   - Scale based on request volume and latency

4. **Troubleshooting**:
   - Check Docker network connectivity
   - Verify backend health endpoints
   - Monitor load balancer metrics
   - Check container logs for errors

## Testing Round Robin Distribution

```bash
bash tests/integration/test_round_robin.sh
```

You should see responses from all active backends in a round-robin pattern.

## Metrics and Monitoring

The load balancer provides detailed metrics via the `/admin/metrics` endpoint:

```bash
curl http://localhost:8080/admin/metrics
```

Key metrics include:
- Request counts per backend
- Response times
- Error rates
- Circuit breaker states
