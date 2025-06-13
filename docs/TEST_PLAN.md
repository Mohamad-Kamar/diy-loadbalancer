# Test Plan

## Core Test Areas

### 1. Backend Management
- Add/remove backends
- Health check verification
- Invalid configurations
- Add duplicate backend URL (should fail)
- Add invalid backend URL (should fail)
- Add backend with incorrect port (should fail)
- Verify backend health check starts after addition
- Verify metrics are created for new backend

### Remove Backend Tests
- Remove existing backend
- Remove non-existent backend (should fail)
- Remove last backend (should maintain minimum backend count)
- Verify metrics cleanup after removal
- Verify in-flight requests complete after removal

## 2. Request Distribution Tests

### Round Robin Verification
- Send sequence of requests
- Verify equal distribution across backends
- Check request headers are preserved
- Validate response consistency
- Monitor distribution metrics

### Load Balancer Features
- Verify circuit breaker activation thresholds
- Test slow backend detection
- Verify request timeouts
- Test concurrent request handling
- Validate backend failure recovery

## 3. Fault Tolerance Tests

### Backend Failure Scenarios
- Simulate backend crash
- Test network latency impact
- Verify circuit breaker behavior
- Test backend recovery detection
- Validate request redistribution

### Resource Limitation Tests
- Test CPU stress impact
- Verify memory limits
- Test high concurrent load
- Validate request queueing
- Monitor resource metrics

## 4. Operational Tests

### Performance Testing
- Measure baseline performance
- Test under normal load
- Test under peak load
- Measure recovery time
- Monitor resource usage

### Configuration Tests
- Verify dynamic config updates
- Test invalid configurations
- Verify config persistence
- Test zero-downtime updates
- Validate logging/metrics

## 5. Integration Tests

### API Conformance
- Verify OpenAPI spec compliance
- Test all HTTP methods
- Validate response formats
- Test error responses
- Verify header handling

### System Integration
- Test all service interactions
- Verify Docker networking
- Test service discovery
- Validate health checks
- Monitor system metrics

## Test Execution

### Prerequisites
1. Docker and Docker Compose installed
2. curl for API testing
3. jq for JSON processing

### Running Tests
```bash
# Start services
docker-compose up -d

# Run test suites
./quick-test.sh      # Quick validation
./test.sh           # Full test suite

# Specific tests
./tests/integration/test_round_robin.sh
./tests/integration/test_phase3.sh
```

### Success Criteria
- All unit tests passing
- Response time < 50ms
- Error rate < 1%
- Even request distribution
- Clean failure handling

### Running Tests
```bash
# 1. Start services
docker-compose up -d

# 2. Run all tests
./test.sh           # Full test suite
./quick-test.sh     # Quick validation

# 3. Run specific test categories
# Unit tests:
cd services/round-robin-api && go test ./internal/... -v  # Round Robin API
cd services/echo-go && go test -v                         # Echo Go Service
cd services/echo-node && npm test                         # Echo Node Service
cd services/echo-java && mvn test                         # Echo Java Service

# Integration tests:
./tests/integration/test_round_robin.sh
./tests/integration/test_phase3.sh

# 4. Cleanup
./cleanup.sh
```

### Monitoring Tests
- Watch Docker logs
- Monitor metrics endpoint
- Check health endpoints
- Review error logs
- Track performance metrics

### Success Criteria
- All unit tests pass
- Integration tests pass
- Performance meets targets
- No memory leaks
- Clean error handling
