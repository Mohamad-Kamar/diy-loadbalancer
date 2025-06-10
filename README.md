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

### Test Round Robin

```bash
bash tests/integration/test_round_robin.sh
```

You should see responses from all three backends in turn.
