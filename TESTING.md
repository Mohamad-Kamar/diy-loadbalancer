# Testing Guide for DIY Load Balancer

A comprehensive guide for testing the load balancer system. All tests passed as of June 13, 2025.

## 🚀 Quick Test

For rapid validation during development:

```bash
./quick-test.sh
```

## 📋 Test Categories

The testing strategy covers three main categories:
1. **Unit Tests** - Individual component testing
2. **Integration Tests** - Cross-service functionality  
3. **Operational Tests** - Real-world scenario validation

**Latest Test Results:** ✅ All tests passed (June 13, 2025) - Production Ready Score: 8.5/10

## 🚀 Quick Start

### Prerequisites
```bash
# Verify required tools are installed
docker --version              # Docker for containerization
docker-compose --version      # Docker Compose for orchestration
go version                    # Go 1.21+ for Go services
node --version                # Node.js 20+ for Node service
java -version && mvn --version # Java 11+ & Maven for Java service
curl --version                # curl for API testing
jq --version                  # jq for JSON parsing (optional but recommended)
```

### Start All Services
```bash
# From project root
docker-compose up --build -d

# Wait for services to initialize
sleep 10

# Verify all services are running
docker-compose ps
```

Expected output:
```
NAME                                STATUS
diy-loadbalancer-round-robin-api-1  Up 
diy-loadbalancer-echo-go-1          Up 
diy-loadbalancer-echo-node-1        Up 
diy-loadbalancer-echo-java-1        Up 
```

## 📋 Test Categories

### 1. Unit Tests (Individual Service Testing)

#### Round Robin API (Go)
```bash
cd services/round-robin-api
go test ./internal/... -v
```
**Tests**: Circuit breaker, health checker, metrics collection, admin API

#### Echo Go Service
```bash
cd services/echo-go
go test -v
```
**Tests**: Echo handler, health endpoint, JSON processing

#### Echo Node Service
```bash
cd services/echo-node
npm test
```
**Tests**: Express endpoints, JSON validation, error handling

#### Echo Java Service
```bash
cd services/echo-java
mvn test
```
**Tests**: HTTP server, echo functionality, concurrent requests

### 2. Integration Tests (End-to-End System Testing)

#### Basic Round Robin Test
```bash
# From project root
./tests/integration/test_round_robin.sh
```
**Validates**: 
- Round-robin distribution across all backends
- JSON echo functionality
- Response consistency

#### Advanced Integration Test
```bash
# From project root
./tests/integration/test_phase3.sh
```
**Validates**:
- Request distribution and metrics
- Backend management (add/remove)
- Error handling scenarios
- Request tracing functionality

### 3. Manual API Tests

#### Health Check
```bash
curl -s http://localhost:8080/admin/health | jq .
```
**Expected**: List of healthy backends with status "ok"

#### Load Balancer API
```bash
curl -s -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"test": "hello"}' | jq .
```
**Expected**: Echo response with same JSON payload

#### Metrics Collection
```bash
curl -s http://localhost:8080/admin/metrics | jq .
```
**Expected**: Request counts, response times, error rates, recent requests

#### Backend Management
```bash
# List backends
curl -s http://localhost:8080/admin/backends | jq .

# Add backend
curl -s -X POST http://localhost:8080/admin/backends \
  -H "Content-Type: application/json" \
  -d '{"url":"http://echo-go:8081"}' | jq .

# Remove backend
curl -s -X DELETE http://localhost:8080/admin/backends \
  -H "Content-Type: application/json" \
  -d '{"url":"http://echo-go:8081"}' | jq .
```

### 4. Load Testing (Optional)

#### Using curl (Simple)
```bash
# Send 30 requests to test distribution
for i in {1..30}; do
  curl -s -X POST http://localhost:8080/api \
    -H "Content-Type: application/json" \
    -d "{\"request\":$i}" | jq -r .request
done | sort | uniq -c
```

#### Using hey (Advanced - if installed)
```bash
# Install hey: go install github.com/rakyll/hey@latest
hey -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"test":"load"}' \
  http://localhost:8080/api
```

### 5. Failure Testing

#### Backend Failure Simulation
```bash
# Stop a backend
docker-compose stop echo-go

# Send requests (should still work with remaining backends)
curl -s -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"test":"failure"}' | jq .

# Restart backend
docker-compose start echo-go
```

#### Network Latency Simulation
```bash
# Add 500ms latency to Node backend
./scripts/add-latency.sh echo-node 500ms

# Test response times
time curl -s -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"test":"latency"}'

# Remove latency
docker exec diy-loadbalancer-echo-node-1 tc qdisc del dev eth0 root netem || true
```

#### CPU Stress Testing
```bash
# Add CPU stress to Java backend
./scripts/cpu-stress.sh echo-java 80%

# Monitor performance
curl -s http://localhost:8080/admin/metrics | jq .response_times

# Stop stress
docker exec diy-loadbalancer-echo-java-1 pkill stress || true
```

## 🎯 Complete Test Suite

### Run All Tests
```bash
# Run the comprehensive test script
./test.sh
```
This script runs:
1. Unit tests for all services
2. Integration tests
3. Basic functionality verification

### Quick Test (Development)
```bash
# Run only tests for changed services
./quick-test.sh
```

## ✅ Test Checklist

Use this checklist to verify all functionality:

### Core Functionality
- [ ] All services start successfully
- [ ] Health endpoints respond correctly
- [ ] Round-robin distribution works
- [ ] JSON echo functionality works
- [ ] All backends receive requests

### Admin API
- [ ] Health endpoint shows all backends
- [ ] Metrics endpoint returns data
- [ ] Backend list endpoint works
- [ ] Add backend functionality works
- [ ] Remove backend functionality works

### Fault Tolerance
- [ ] System works with one backend down
- [ ] Circuit breaker activates on failures
- [ ] Backends recover after restart
- [ ] Timeouts handled correctly
- [ ] Invalid requests rejected properly

### Performance
- [ ] Response times are reasonable
- [ ] System handles concurrent requests
- [ ] Memory usage is stable
- [ ] Connection pooling works

## 🔧 Troubleshooting

### Services Won't Start
```bash
# Check logs
docker-compose logs

# Rebuild and restart
docker-compose down
docker-compose up --build -d
```

### Tests Failing
```bash
# Check service health
docker-compose ps

# Check individual service logs
docker-compose logs [service-name]

# Restart specific service
docker-compose restart [service-name]
```

### Port Conflicts
```bash
# Check what's using ports
lsof -i :8080
lsof -i :8081
lsof -i :8082
lsof -i :8083

# Stop conflicting processes or change ports in docker-compose.yml
```

## 📊 Expected Test Results

### Unit Tests
- All Go tests: PASS
- Node tests: All specs passing
- Java tests: All JUnit tests green

### Integration Tests
- Round robin distribution: Even distribution ±2 requests
- Backend management: Add/remove operations successful
- Error handling: Proper HTTP status codes
- Request tracing: Request IDs in metrics

### Performance Baseline
- Average response time: < 50ms
- P95 response time: < 100ms
- Error rate: < 1%
- Memory usage: < 512MB total for all services
