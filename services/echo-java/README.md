# Echo Java Service

A simple HTTP echo service built with Java 11 using only standard library components (no external dependencies like Spring).

## Features

- **POST `/`**: Echoes back the JSON payload sent in the request body
- **GET `/health`**: Returns a health check response `{"status":"ok"}`
- **Port**: Runs on port 8083

## Prerequisites

- Java 11 or higher
- Maven 3.6+ 

## Quick Start

### 1. Build the application
```bash
mvn clean package
```

### 2. Run the application
```bash
# Option 1: Run with Maven
mvn exec:java

# Option 2: Run the JAR directly
java -jar target/echo-java-0.0.1-SNAPSHOT.jar
```

### 3. Test the service

**Automated Test (Recommended):**
```bash
# Run the test script (service must be running)
./test-java-service.sh
```

**Manual Tests:**

**Health Check:**
```bash
curl http://localhost:8083/health
# Expected: {"status":"ok"}
```

**Echo Test:**
```bash
curl -X POST http://localhost:8083/ \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello from Java!", "timestamp": "2024-01-01"}'
# Expected: {"message": "Hello from Java!", "timestamp": "2024-01-01"}
```

## Maven Commands

| Command | Description |
|---------|-------------|
| `mvn clean` | Clean build artifacts |
| `mvn compile` | Compile source code |
| `mvn package` | Build JAR file |
| `mvn clean package` | Clean and build |
| `mvn exec:java` | Run application directly |
| `mvn clean package exec:java` | Build and run in one command |


```bash
# 1. First time setup - build the project
mvn clean package

# 2. Run the application
mvn exec:java

# 3. In another terminal, test it
./test-java-service.sh

# 4. Make changes to code, then rebuild and run
mvn clean package exec:java
```

## Docker Support

Build and run with Docker:
```bash
docker build -t echo-java .
docker run -p 8083:8083 echo-java
```

## Architecture Notes

- Uses Java's built-in `com.sun.net.httpserver.HttpServer` 
- No external dependencies for maximum simplicity
- Fixed thread pool with 10 threads for handling requests
