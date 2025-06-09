## Executive Summary

This assignment is an opportunity to demonstrate **distributed systems expertise** beyond basic load balancing. The focus should be on building a **resilient, observable foundation** that handles real-world failure scenarios while showcasing custom implementation skills.

## Requirements Analysis & Research

### Core Requirements
1. **Application API**: HTTP POST echo service with JSON payload
2. **Round Robin API**: Custom load balancer with configurable backends  
3. **Multi-instance Support**: Minimum 3 backend instances
4. **Fault Tolerance**: Handle backend failures and slowness
5. **Testing Strategy**: Comprehensive validation approach

### Enhanced Requirements (From Notes)
- **Polyglot Backends**: Java, Go, JavaScript for performance diversity
- **Resource Management**: Hardware limitations per service  
- **Dynamic Configuration**: Runtime backend management
- **Observability Foundation**: Request counting and distribution metrics
- **Independent Deployment**: Containerized services

## Industry Research: How Production Systems Solve These Problems

### **Load Balancer Patterns**
- **NGINX/HAProxy**: Active health checks, weighted round-robin, connection pooling
- **AWS ALB/GCP LB**: Target group health monitoring, automatic failover
- **Envoy Proxy**: Circuit breakers, outlier detection, observability integration

### **Key Patterns to Implement**
1. **Health Check Endpoints**: Dedicated `/health` routes for monitoring
2. **Circuit Breaker Pattern**: Prevent cascade failures from slow/failed backends  
3. **Request Timeout Management**: Prevent resource exhaustion
4. **Graceful Degradation**: Continue serving when subset of backends fail

## Project Architecture

### High-Level Design
```
┌─────────────┐    POST /api     ┌──────────────────┐
│   Client    │ ──────────────▶  │ Round Robin API  │
└─────────────┘                  │ (Go)             │
                                 └─────────┬────────┘
                                           │
                        ┌──────────────────┼──────────────────┐
                        ▼                  ▼                  ▼
                ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
                │ App API      │  │ App API      │  │ App API      │
                │ (Go:8081)    │  │ (Node:8082)  │  │ (Java:8083)  │
                │ CPU: 1.0     │  │ CPU: 0.7     │  │ CPU: 0.5     │
                │ RAM: 128MB   │  │ RAM: 256MB   │  │ RAM: 512MB   │
                └──────────────┘  └──────────────┘  └──────────────┘
```

### Component Specifications

**Application APIs (Echo Services)**
- **Go Service**: Lightweight, fast response (~5ms)
- **Node.js Service**: Medium performance (~15ms) 
- **Java Service**: Heavier, slower startup (~25ms)
- Each exposes: `POST /` (echo) and `GET /health` (monitoring)

**Round Robin API (Go)**
```go
type LoadBalancer struct {
    backends    []Backend
    currentIdx  uint64
    healthCheck *HealthChecker
    metrics     *MetricsCollector
    circuit     map[string]*CircuitBreaker
}

type Backend struct {
    URL         string
    Healthy     bool
    Weight      int
    RequestCount uint64
    AvgLatency   time.Duration
}
```

## Repository Strategy: Monorepo with Independent Deployment

### Recommended Structure
```
round-robin-api/
├── services/
│   ├── round-robin-api/        # Go load balancer
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── balancer/
│   │   │   ├── health/
│   │   │   └── metrics/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── echo-go/               # Go echo service
│   ├── echo-node/             # Node.js echo service  
│   └── echo-java/             # Java echo service
├── deployment/
│   ├── docker-compose.yml
│   ├── k8s/                   # Kubernetes manifests
│   └── scripts/
├── tests/
│   ├── integration/
│   └── load/
└── docs/
    ├── api-spec.yml           # OpenAPI specification
    └── architecture.md
```

**Rationale for Monorepo:**
- **Simplified Development**: Single clone, unified CI/CD
- **Atomic Changes**: Cross-service updates in single commit
- **Shared Tooling**: Common Docker, testing, documentation
- **Demo Friendly**: Easy to package and present

**Independent Deployment Strategy:**
```yaml
# CI/CD Pipeline (conceptual)
services:
  - name: round-robin-api
    path: services/round-robin-api/
    image: registry/round-robin-api
  - name: echo-go  
    path: services/echo-go/
    image: registry/echo-go
```

## Phased Execution Plan

### **Phase 1: Core MVP** (4-6 hours)
**Priority: Basic Functionality**

#### Deliverables:
1. **Echo Services Implementation**
   ```go
   // Go service
   func echoHandler(w http.ResponseWriter, r *http.Request) {
       var payload map[string]interface{}
       json.NewDecoder(r.Body).Decode(&payload)
       json.NewEncoder(w).Encode(payload)
   }
   ```

2. **Basic Round Robin Logic**
   ```go
   func (lb *LoadBalancer) NextBackend() *Backend {
       if len(lb.backends) == 0 {
           return nil
       }
       idx := atomic.AddUint64(&lb.currentIdx, 1) % uint64(len(lb.backends))
       return &lb.backends[idx]
   }
   ```

3. **Docker Infrastructure**
   ```yaml
   version: '3.8'
   services:
     round-robin-api:
       build: ./services/round-robin-api
       ports: ["8080:8080"]
       environment:
         - BACKENDS=http://echo-go:8081,http://echo-node:8082,http://echo-java:8083
     
     echo-go:
       build: ./services/echo-go
       deploy:
         resources:
           limits: {cpus: '1.0', memory: '128M'}
   ```

**Success Criteria:**
- All services start via `docker-compose up`
- Round-robin distribution works across 3 backends
- JSON echo functionality verified

### **Phase 2: Resilience & Health Monitoring** (3-4 hours)
**Priority: Production Readiness**

#### Deliverables:
1. **Active Health Checks**
   ```go
   func (hc *HealthChecker) monitorBackend(backend *Backend) {
       ticker := time.NewTicker(10 * time.Second)
       for range ticker.C {
           if hc.checkHealth(backend.URL + "/health") {
               backend.MarkHealthy()
           } else {
               backend.MarkUnhealthy()
           }
       }
   }
   ```

2. **Circuit Breaker Implementation**
   ```go
   type CircuitBreaker struct {
       state            State // CLOSED, OPEN, HALF_OPEN
       failureThreshold int
       failureCount     int
       lastFailureTime  time.Time
       timeout          time.Duration
   }
   ```

3. **Graceful Failure Handling**
   - Automatic backend exclusion when unhealthy
   - Request timeout handling (2s default)
   - Fallback responses when all backends down

**Success Criteria:**
- System continues working when backend stopped
- Failed backends automatically recovered when restarted
- Circuit breaker protects against slow backends

### **Phase 3: Observability & Operations** (2-3 hours)
**Priority: Operational Excellence**

#### Deliverables:
1. **Metrics Collection**
   ```go
   type Metrics struct {
       RequestCounts    map[string]uint64
       ResponseTimes    map[string]time.Duration
       ErrorRates       map[string]float64
       CircuitStates    map[string]State
   }
   ```

2. **Admin API**
   ```http
   GET  /admin/metrics          # Request distribution
   GET  /admin/health           # Backend status  
   POST /admin/backends         # Add backend
   DELETE /admin/backends/{url} # Remove backend
   ```

3. **Dynamic Configuration**
   - Runtime backend addition/removal
   - Configuration reload without restart
   - Environment variable and API-based config

**Success Criteria:**
- Metrics show request distribution
- Dynamic backend management works
- Zero-downtime configuration changes

### **Phase 4: Testing & Excellence** (2-3 hours)
**Priority: Quality Assurance**

#### Deliverables:
1. **Comprehensive Test Suite**
   ```bash
   # Unit tests
   go test ./internal/balancer -v
   
   # Integration tests  
   ./scripts/integration-test.sh
   
   # Load tests
   ./scripts/load-test.sh
   ```

2. **Failure Simulation**
   ```bash
   # Simulate backend failure
   docker-compose stop echo-java
   
   # Simulate slow backend
   ./scripts/add-latency.sh echo-node 500ms
   
   # Simulate CPU load
   ./scripts/cpu-stress.sh echo-java 80%
   ```

3. **Documentation**
   - Architecture decision records
   - API documentation (OpenAPI)
   - Deployment guides
   - Demo scripts

## Essential Technical Implementation

### **1. Core Round Robin Algorithm**
```go
type RoundRobinBalancer struct {
    backends []Backend
    index    uint64
    mutex    sync.RWMutex
}

func (rr *RoundRobinBalancer) Next() (*Backend, error) {
    rr.mutex.RLock()
    healthy := rr.getHealthyBackends()
    rr.mutex.RUnlock()
    
    if len(healthy) == 0 {
        return nil, ErrNoHealthyBackends
    }
    
    idx := atomic.AddUint64(&rr.index, 1) % uint64(len(healthy))
    return &healthy[idx], nil
}
```

### **2. Health Check System**
```go
func (hc *HealthChecker) CheckBackend(ctx context.Context, backend *Backend) error {
    client := &http.Client{Timeout: 5 * time.Second}
    
    resp, err := client.Get(backend.URL + "/health")
    if err != nil || resp.StatusCode != 200 {
        backend.MarkUnhealthy()
        return err
    }
    
    backend.MarkHealthy()
    return nil
}
```

### **3. Configuration Management**
```go
type Config struct {
    Backends []BackendConfig `json:"backends" env:"BACKENDS"`
    Health   HealthConfig    `json:"health"`
    Timeouts TimeoutConfig   `json:"timeouts"`
}

func LoadConfig() (*Config, error) {
    // Priority: CLI args > env vars > config file > defaults
}
```

## Testing Strategy

### **Unit Tests**
- Round robin index calculation
- Circuit breaker state transitions  
- Health check logic
- Configuration parsing

### **Integration Tests**
```bash
#!/bin/bash
# integration-test.sh

# Start services
docker-compose up -d

# Wait for health
wait_for_health() {
    curl -f http://localhost:8080/admin/health || sleep 2
}

# Test round robin distribution
test_distribution() {
    for i in {1..30}; do
        curl -X POST http://localhost:8080/api \
             -H "Content-Type: application/json" \
             -d '{"test": "data"}'
    done
    
    # Verify distribution via metrics
    curl http://localhost:8080/admin/metrics | jq '.request_counts'
}
```

### **Load Tests**
```bash
# Use hey or ab for load testing
hey -n 1000 -c 10 -m POST \
    -H "Content-Type: application/json" \
    -d '{"load": "test"}' \
    http://localhost:8080/api
```

### **Chaos Testing**
```bash
# Simulate network partition
docker network disconnect round-robin-api_default echo-java

# Simulate memory pressure  
docker exec echo-node stress --vm 1 --vm-bytes 200M --timeout 60s

# Simulate CPU load
docker exec echo-go stress --cpu 2 --timeout 30s
```

## Resource Management & Performance

### **Container Resource Limits**
```yaml
services:
  echo-java:
    deploy:
      resources:
        limits:
          cpus: '0.5'      # Slower service
          memory: '512M'
        reservations:
          cpus: '0.25'
          memory: '256M'
          
  echo-go:  
    deploy:
      resources:
        limits:
          cpus: '1.0'      # Fastest service
          memory: '128M'
```

### **Performance Monitoring**
```go
type PerformanceMetrics struct {
    RequestLatency   map[string][]time.Duration
    ThroughputRPS    map[string]float64
    ErrorRate        map[string]float64
    ResourceUsage    map[string]ResourceStats
}
```

## Excellence Considerations

### **Missing Points Often Overlooked**

1. **Request Correlation**
   ```go
   type RequestContext struct {
       ID          string    // UUID for tracing
       StartTime   time.Time
       BackendURL  string
       UserAgent   string
   }
   ```

2. **Graceful Shutdown**
   ```go
   func (s *Server) Shutdown(ctx context.Context) error {
       // Drain existing connections
       // Stop accepting new requests
       // Wait for inflight requests
       return s.httpServer.Shutdown(ctx)
   }
   ```

3. **Security Considerations**
   - Request size limits (prevent DoS)
   - Input validation and sanitization
   - Rate limiting per client
   - CORS configuration

4. **Operational Readiness**
   - Structured logging with levels
   - Health check endpoints for monitoring
   - Prometheus metrics export
   - Docker health checks

### **Future Extensibility**

1. **Pluggable Algorithms**
   ```go
   type LoadBalancer interface {
       Next() (*Backend, error)
       AddBackend(backend Backend) error
       RemoveBackend(url string) error
   }
   
   // Implementations: RoundRobin, WeightedRoundRobin, LeastConnections
   ```

2. **Service Discovery Integration**
   ```go
   type ServiceDiscovery interface {
       Watch(serviceName string) (<-chan []Backend, error)
       Register(service Service) error
   }
   ```

3. **Advanced Monitoring**
   - Distributed tracing integration
   - Custom dashboards (Grafana)
   - Alert rules (Prometheus)

## Deployment Strategies

### **Local Development**
```bash
# One-command startup
docker-compose up --build

# Scale specific service
docker-compose up --scale echo-go=3

# Rolling updates
docker-compose up -d --no-deps echo-java
```

### **Production Deployment**
```yaml
# Kubernetes example
apiVersion: apps/v1
kind: Deployment
metadata:
  name: round-robin-api
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  template:
    spec:
      containers:
      - name: round-robin-api
        image: registry/round-robin-api:v1.2.3
        resources:
          requests: {memory: "256Mi", cpu: "250m"}
          limits: {memory: "512Mi", cpu: "500m"}
```

## Risk Mitigation

### **Technical Risks**
- **Single Point of Failure**: Deploy multiple load balancer instances
- **Memory Leaks**: Implement connection pooling and proper cleanup
- **Configuration Drift**: Version control all configuration
- **Performance Degradation**: Continuous load testing in CI/CD

### **Operational Risks**  
- **Deployment Issues**: Blue-green deployment strategy
- **Monitoring Gaps**: Comprehensive health checks and alerts
- **Security Vulnerabilities**: Regular dependency scanning
- **Data Loss**: Stateless design with external state storage

## Summary

This approach transforms a basic coding assignment into a **comprehensive distributed systems showcase** by:

1. **Implementing proven patterns** from production load balancers
2. **Demonstrating operational thinking** through monitoring and failure handling
3. **Building extensible foundations** for future enhancements
4. **Providing comprehensive testing** including chaos engineering
5. **Creating production-ready deployment** strategies

The phased approach ensures **incremental value delivery** while building toward a **an implementation** that showcases both coding skills and systems thinking.

**Key Success Metrics:**
- ✅ Custom round-robin logic implemented
- ✅ Handles backend failures gracefully  
- ✅ Polyglot backend services with resource constraints
- ✅ Dynamic configuration capability
- ✅ Comprehensive observability foundation
- ✅ Production-ready deployment strategy
