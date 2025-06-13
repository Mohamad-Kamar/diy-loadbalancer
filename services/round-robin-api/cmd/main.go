package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"round-robin-api/internal/admin"
	"round-robin-api/internal/circuit"
	"round-robin-api/internal/logger"
	"round-robin-api/internal/metrics"
)

type Backend struct {
	URL       string
	breaker   *circuit.CircuitBreaker
	isHealthy bool
}

type LoadBalancer struct {
	sync.RWMutex
	Backends      []Backend
	currIndex    uint64
	healthChecker *circuit.HealthChecker
	metrics      *metrics.Metrics
	client       *http.Client
	logger       *logger.Logger
}

func NewLoadBalancer(urls []string) *LoadBalancer {
	backends := make([]Backend, len(urls))
	healthChecker := circuit.NewHealthChecker()
	metricsCollector := metrics.NewMetrics()
	appLogger := logger.New(logger.INFO)

	for i, u := range urls {
		backends[i] = Backend{
			URL:       u,
			breaker:   circuit.NewCircuitBreaker(),
			isHealthy: true,
		}
		// Start health checking for this backend
		healthChecker.StartChecking(u, time.Second*5) // Check every 5 seconds
		appLogger.Info("Added backend: %s", u)
	}

	return &LoadBalancer{
		Backends:      backends,
		healthChecker: healthChecker,
		metrics:      metricsCollector,
		logger:       appLogger,
		client: &http.Client{
			Timeout: time.Second * 2, // 2 second timeout for requests
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}
}

// AddBackend adds a new backend to the load balancer
func (lb *LoadBalancer) AddBackend(url string) {
	lb.Lock()
	defer lb.Unlock()

	// Check if backend already exists
	for _, b := range lb.Backends {
		if b.URL == url { // URL is already normalized by admin layer
			lb.logger.Warn("Backend already exists: %s", url)
			return // Already exists
		}
	}

	backend := Backend{
		URL:       url,
		breaker:   circuit.NewCircuitBreaker(),
		isHealthy: true,
	}
	lb.Backends = append(lb.Backends, backend)
	lb.healthChecker.StartChecking(url, time.Second*5)
	lb.logger.Info("Added new backend: %s", url)
}

// RemoveBackend removes a backend from the load balancer
func (lb *LoadBalancer) RemoveBackend(url string) {
	lb.Lock()
	defer lb.Unlock()

	initialCount := len(lb.Backends)
	newBackends := make([]Backend, 0)
	for _, b := range lb.Backends {
		if b.URL != url { // URL is already normalized by admin layer
			newBackends = append(newBackends, b)
		}
	}
	lb.Backends = newBackends
	
	if len(lb.Backends) < initialCount {
		lb.logger.Info("Removed backend: %s", url)
	} else {
		lb.logger.Warn("Backend not found for removal: %s", url)
	}
}

// GetBackends returns a list of backend URLs
func (lb *LoadBalancer) GetBackends() []string {
	lb.RLock()
	defer lb.RUnlock()

	urls := make([]string, len(lb.Backends))
	for i, backend := range lb.Backends {
		urls[i] = backend.URL
	}
	return urls
}

func (lb *LoadBalancer) NextBackend() *Backend {
	lb.RLock()
	defer lb.RUnlock()

	if len(lb.Backends) == 0 {
		return nil
	}

	// Try up to N times to find a healthy backend
	maxAttempts := len(lb.Backends)
	for i := 0; i < maxAttempts; i++ {
		idx := atomic.AddUint64(&lb.currIndex, 1) % uint64(len(lb.Backends))
		backend := &lb.Backends[idx]
		
		// Check if backend is healthy and circuit is available
		if lb.healthChecker.IsHealthy(backend.URL) && backend.breaker.IsAvailable() {
			// Record metrics
			lb.metrics.RecordRequest(backend.URL)
			lb.metrics.RecordCircuitState(backend.URL, backend.breaker.GetState())
			return backend
		}
	}
	return nil
}

func (lb *LoadBalancer) forwardRequest(backend *Backend, r *http.Request) (*http.Response, error) {
	start := time.Now()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	// Create new request with timeout context
	req, err := http.NewRequestWithContext(ctx, r.Method, backend.URL+"/", r.Body)
	if err != nil {
		lb.metrics.RecordRequestComplete(r.Header.Get("X-Request-ID"), backend.URL, time.Since(start), false)
		return nil, err
	}

	// Copy headers
	for key, values := range r.Header {
		req.Header[key] = values
	}

	// Add or update request ID
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		req.Header.Set("X-Request-ID", requestID)
	}

	// Forward the request
	resp, err := lb.client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		backend.breaker.RecordFailure()
		lb.metrics.RecordRequestComplete(requestID, backend.URL, duration, false)
		return nil, err
	}

	success := resp.StatusCode < 500
	if success {
		backend.breaker.RecordSuccess()
	} else {
		backend.breaker.RecordFailure()
	}

	lb.metrics.RecordRequestComplete(requestID, backend.URL, duration, success)
	return resp, nil
}

func main() {
	appLogger := logger.New(logger.INFO)
	
	backendEnv := os.Getenv("BACKENDS")
	if backendEnv == "" {
		appLogger.Fatal("BACKENDS env var required")
	}
	urls := strings.Split(backendEnv, ",")
	appLogger.Info("Starting load balancer with %d backends", len(urls))
	
	lb := NewLoadBalancer(urls)

	// Create admin server
	adminServer := admin.NewAdminServer(lb.metrics, lb)

	// Main API endpoint
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		// Add request ID for tracing
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		contextLogger := lb.logger.WithRequestID(requestID)
		
		contextLogger.Debug("Received request: %s %s", r.Method, r.URL.Path)

		if r.Method != http.MethodPost {
			contextLogger.Warn("Method not allowed: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method not allowed"}`))
			return
		}

		// Limit request body size (1MB max)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		// Validate content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			contextLogger.Warn("Invalid content type: %s", contentType)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Content-Type must be application/json"}`))
			return
		}

		backend := lb.NextBackend()
		if backend == nil {
			contextLogger.Error("No healthy backends available")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"No healthy backends available"}`))
			return
		}

		contextLogger.Debug("Forwarding to backend: %s", backend.URL)
		
		resp, err := lb.forwardRequest(backend, r)
		if err != nil {
			if err == context.DeadlineExceeded {
				contextLogger.Error("Backend timeout: %s", backend.URL)
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte(`{"error":"Backend timeout"}`))
				return
			}
			contextLogger.Error("Backend error: %s - %v", backend.URL, err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error":"Backend error"}`))
			return
		}
		defer resp.Body.Close()

		// Add X-Served-By header for debugging/testing
		w.Header().Set("X-Served-By", backend.URL)
		w.Header().Set("X-Request-ID", requestID)
		
		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		
		contextLogger.Debug("Request completed successfully")
	})

	// Admin API endpoints
	http.HandleFunc("/admin/metrics", adminServer.HandleMetrics)
	http.HandleFunc("/admin/health", adminServer.HandleHealth)
	http.HandleFunc("/admin/backends", adminServer.HandleBackends)

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		appLogger.Info("Round Robin API listening on :8080")
		appLogger.Info("Admin API available at /admin/*")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Server failed to start: %v", err)
		}
	}()

	<-stop
	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown: %v", err)
	} else {
		appLogger.Info("Server gracefully stopped")
	}
}
