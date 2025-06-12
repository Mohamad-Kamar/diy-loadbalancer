# diy-loadbalancer

## Overview
Load balancer and echo service platform. This project demonstrates distributed systems best practices: resilience, observability, and dynamic configuration. It features a custom round-robin load balancer (Go) and three echo backends (Go, Node.js, Java), all containerized and ready for real-world failure scenarios.

---

## Features
- **Custom Round-Robin Load Balancer** (Go)
- **Polyglot Echo Backends**: Go, Node.js, Java
- **Dynamic Backend Management**: Add/remove backends at runtime via Admin API
- **Fault Tolerance**: Health checks, circuit breaker, request timeouts
- **Observability**: Metrics, request tracing, distribution stats
- **Zero-Downtime Config Reload**
- **Comprehensive Test Suite**: Unit, integration, and operational tests
- **Failure Simulation Scripts**: Latency and CPU stress injection
- **Containerized Deployment**: Docker Compose for easy orchestration

---

## Architecture
```
[Client] ⇄ [Load Balancer (Go)] ⇄ [Echo Backends: Go | Node.js | Java]
```
- **/api**: Main POST endpoint, round-robin to healthy backends
- **/admin/**: Admin API for metrics, health, and backend management

---

## Quick Start

### Option 1: Docker Compose (Recommended)
```bash
docker-compose up --build
```
- Load balancer: http://localhost:8080
- Backends: http://localhost:8081 (Go), :8082 (Node), :8083 (Java)

### Option 2: Manual (Development)
Start each backend in its directory:
```bash
# Go backend
go run main.go
# Node backend
node index.js
# Java backend
mvn exec:java
```
Start the load balancer:
```bash
cd services/round-robin-api
./run.sh
# or
BACKENDS=http://localhost:8081,http://localhost:8082,http://localhost:8083 go run cmd/main.go
```

---

## API Reference
See [`docs/openapi.yaml`](docs/openapi.yaml) for full OpenAPI spec.

### Main Endpoints
- `POST /api` — Echo POST, round-robin to healthy backend
- `GET /admin/metrics` — Metrics and request distribution
- `GET /admin/health` — Backend health status
- `GET /admin/backends` — List backends
- `POST /admin/backends` — Add backend (`{"url": "http://..."}`)
- `DELETE /admin/backends` — Remove backend (`{"url": "http://..."}`)

---

## Dynamic Backend Management
- Add/remove backends at runtime via Admin API (see above)
- No restart required; config reloads live

---

## Observability & Metrics
- Request counts, error rates, response times per backend
- Circuit breaker state
- Recent request log
- All available at `/admin/metrics`

---

## Testing & Quality Assurance

### Unit Tests
- **Go**: `go test ./...` in each Go service directory
- **Node.js**: `npm test` (uses Jest & Supertest)
- **Java**: `mvn test` (JUnit 5)

### Integration & Operational Tests
- Run all integration tests:
  ```bash
  ./scripts/run-all-tests.sh
  ```
- See `tests/integration/` for round-robin, failure, and observability tests

### Failure Simulation
- **Add latency:**
  ```bash
  ./scripts/add-latency.sh <container_name> <latency>
  # Example: ./scripts/add-latency.sh echo-node 500ms
  ```
- **CPU stress:**
  ```bash
  ./scripts/cpu-stress.sh <container_name> <percent>
  # Example: ./scripts/cpu-stress.sh echo-java 80%
  ```
- Remove latency: `docker exec <container> tc qdisc del dev eth0 root netem || true`
- Stop stress: `docker exec <container> pkill stress || true`

---

## Testing Guide

### Running All Tests
```bash
# Run all tests (unit tests + integration tests)
./test.sh

# Run quick tests (only for changed services)
./quick-test.sh

# Clean up all build artifacts and containers
./cleanup.sh
```

### Test Categories

#### 1. Unit Tests
- **Round Robin API (Go)**:
  - Load balancer logic
  - Circuit breaker behavior
  - Health check system
  - Backend management
  
- **Echo Services**:
  - Request handling
  - Health endpoint
  - Error scenarios

#### 2. Integration Tests
- **Backend Management**:
  - Adding/removing backends dynamically
  - Validation of backend URLs
  - Health status updates

- **Request Distribution**:
  - Round robin distribution verification
  - Headers and payload forwarding
  - Response consistency

- **Fault Tolerance**:
  - Circuit breaker activation
  - Recovery from backend failures
  - Timeout handling
  - Load shedding

#### 3. Operational Tests
- **Performance**:
  - Load testing (using hey or ab)
  - Response time measurements
  - Resource utilization

- **Failure Simulation**:
  ```bash
  # Add latency to a backend
  ./scripts/add-latency.sh echo-node 500ms

  # Simulate CPU load
  ./scripts/cpu-stress.sh echo-java 80%

  # Stop a backend
  docker-compose stop echo-go

  # Start a backend
  docker-compose start echo-go
  ```

### Test Coverage Goals
- Unit test coverage: >80%
- Integration test scenarios: >90%
- All critical paths tested
- Fault injection for all failure modes

### Continuous Testing
The test suite is designed to run in CI/CD pipelines:
```bash
# Build all services
docker-compose build

# Run test suite
./scripts/run-all-tests.sh

# Clean up
docker-compose down
```

---

## Project Structure
```
services/
  echo-go/      # Go echo backend
  echo-node/    # Node.js echo backend
  echo-java/    # Java echo backend
  round-robin-api/ # Go load balancer
scripts/        # Test and simulation scripts
tests/          # Integration tests
docs/           # OpenAPI, ADRs, architecture docs
```

---

## Contributing
- Fork and PRs welcome!
- See `docs/adr-template.md` for architecture decision records
- Please add/extend unit and integration tests for new features

---

## Authors & Credits
- Kamar
- Inspired by industry best practices (NGINX, Envoy, AWS ALB, etc.)

---

## License
MIT
