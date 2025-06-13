# DIY Load Balancer

A production-ready round-robin load balancer implementation showcasing distributed systems best practices: resilience, observability, and dynamic configuration.

[![Production Ready](https://img.shields.io/badge/Production%20Ready-8.5%2F10-brightgreen)](#production-readiness)
[![Tests](https://img.shields.io/badge/Tests-All%20Passing-brightgreen)](#testing)
[![Polyglot](https://img.shields.io/badge/Backends-Go%2BNode.js%2BJava-blue)](#architecture)

## 🏗️ Architecture

```
┌─────────────┐    ┌─────────────────────┐    ┌─────────────────────────┐
│   Clients   │────│  Load Balancer (Go) │────│   Echo Backends         │
│             │    │                     │    │                         │
│ HTTP Clients│◄───│ • Round-Robin       │◄───│ • Go Service    :8081   │
│             │    │ • Health Checks     │    │ • Node.js Service :8082 │
│             │    │ • Circuit Breaker   │    │ • Java Service   :8083  │
│             │    │ • Admin API         │    │                         │
└─────────────┘    └─────────────────────┘    └─────────────────────────┘
```

**Core Components:**
- **Load Balancer (Go)**: Custom round-robin implementation with health monitoring
- **Polyglot Backends**: Three identical echo services in different languages
- **Admin API**: Runtime backend management and observability

## ✨ Features

### Load Balancing
- ✅ **Round-Robin Distribution**: Even traffic distribution across healthy backends
- ✅ **Health Monitoring**: Automatic failure detection and recovery
- ✅ **Circuit Breaker**: Prevents cascading failures
- ✅ **Request Timeouts**: Configurable per-request timeouts

### Observability
- ✅ **Metrics Collection**: Request counts, response times, error rates
- ✅ **Request Tracing**: Unique request IDs for debugging
- ✅ **Structured Logging**: JSON-formatted logs with context
- ✅ **Health Endpoints**: Real-time system status

### Operations
- ✅ **Dynamic Configuration**: Add/remove backends at runtime
- ✅ **Graceful Shutdown**: Drain connections before stopping
- ✅ **Zero-Downtime Updates**: Backend changes without service interruption
- ✅ **Connection Pooling**: Optimized HTTP client performance

### Quality Assurance
- ✅ **Comprehensive Testing**: Unit, integration, and operational tests
- ✅ **Failure Simulation**: Network latency and CPU stress injection
- ✅ **Production Validation**: Real-world scenario testing

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for development)
- Node.js 20+ (for development)
- Java 11+ & Maven (for development)

### Start All Services
```bash
# Clone and navigate to project
git clone <repository-url>
cd diy-loadbalancer

# Start all services with Docker Compose
docker-compose up --build -d

# Verify services are running
docker-compose ps
```

**Service Endpoints:**
- Load Balancer: http://localhost:8080
- Admin API: http://localhost:8080/admin/*
- Go Backend: http://localhost:8081
- Node.js Backend: http://localhost:8082  
- Java Backend: http://localhost:8083

### Test the System
```bash
# Quick health check
curl http://localhost:8080/admin/health

# Send test requests
curl -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello World"}'

# Check metrics
curl http://localhost:8080/admin/metrics
## 🧪 Testing

### Quick Validation (4 minutes)
```bash
# Run comprehensive test suite
./quick-test.sh

# Or follow step-by-step guide in TESTING.md
```

### Test Categories
1. **Health Checks** - Verify all services are operational
2. **Load Balancing** - Confirm round-robin distribution  
3. **Failure Handling** - Test backend failure and recovery
4. **Admin API** - Validate runtime configuration
5. **Error Handling** - Check proper HTTP responses

**Latest Test Results:** ✅ All tests passed (June 13, 2025)

## 📡 API Reference

### Main API

#### POST /api
Echo endpoint that forwards requests to backend services.

**Request:**
```bash
curl -X POST http://localhost:8080/api \
  -H "Content-Type: application/json" \
  -d '{"key": "value"}'
```

**Response:**
```json
{
  "key": "value"
}
```

**Headers:**
- `X-Served-By`: Backend that handled the request
- `X-Request-ID`: Unique request identifier

### Admin API

#### GET /admin/health
System health status and backend list.

#### GET /admin/metrics  
Comprehensive system metrics including:
- Request counts per backend
- Response times and error rates
- Circuit breaker states
- Recent request history

#### GET /admin/backends
List all configured backends.

#### POST /admin/backends
Add a new backend dynamically.

**Request:**
```json
{
  "url": "http://new-backend:8084"
}
```

#### DELETE /admin/backends
Remove a backend from rotation.

**Request:**
```json
{
  "url": "http://backend-to-remove:8084"
}
```

## 🔧 Development

### Project Structure
```
diy-loadbalancer/
├── services/
│   ├── round-robin-api/    # Load balancer (Go)
│   ├── echo-go/            # Go backend service
│   ├── echo-node/          # Node.js backend service
│   └── echo-java/          # Java backend service
├── tests/                  # Integration tests
├── scripts/                # Utility scripts
└── TESTING.md             # Complete testing guide
```

### Running Individual Services

**Load Balancer:**
```bash
cd services/round-robin-api
export BACKENDS=http://localhost:8081,http://localhost:8082,http://localhost:8083
go run cmd/main.go
```

**Go Backend:**
```bash
cd services/echo-go
go run main.go
```

**Node.js Backend:**
```bash
cd services/echo-node
npm install
npm start
```

**Java Backend:**
```bash
cd services/echo-java
mvn clean package
mvn exec:java
```

### Testing During Development
```bash
# Run unit tests for all services
./test.sh

# Run integration tests
./tests/integration/test_round_robin.sh

# Simulate failure scenarios
./scripts/add-latency.sh echo-node 500ms
./scripts/cpu-stress.sh echo-java 80%
```

## 🏆 Production Readiness

**Current Score: 8.5/10**

### ✅ Implemented Features
- Load balancing with health checks
- Fault tolerance and recovery
- Comprehensive observability
- Dynamic configuration
- Graceful shutdown
- Connection optimization
- Complete test coverage

### 🔄 Future Enhancements
- SSL/TLS termination
- Rate limiting and throttling
- Advanced load balancing algorithms
- Monitoring/alerting integration
- Kubernetes deployment manifests

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature-name`
3. **Run tests**: `./test.sh`
4. **Commit changes**: `git commit -am 'Add feature'`
5. **Push to branch**: `git push origin feature-name`
6. **Submit a Pull Request**

### Code Standards
- Go: `gofmt`, `go vet`, comprehensive tests
- Node.js: ESLint, Jest testing
- Java: Maven standards, JUnit tests
- All services: Dockerfile optimization

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [`README.md`](README.md) | Main project overview and quick start guide |
| [`TESTING.md`](TESTING.md) | Comprehensive testing guide and procedures |
| [`docs/DESIGN.md`](docs/DESIGN.md) | Architecture decisions and project planning |
| [`docs/TEST_PLAN.md`](docs/TEST_PLAN.md) | Detailed test scenarios and validation |
| [`docs/openapi.yaml`](docs/openapi.yaml) | OpenAPI specification for all APIs |

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built as a demonstration of distributed systems engineering principles, showcasing:
- Circuit breaker pattern implementation
- Health monitoring strategies  
- Polyglot service architecture
- Production-ready operational practices

---