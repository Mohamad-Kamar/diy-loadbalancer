package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"round-robin-api/internal/admin"
	"round-robin-api/internal/circuit"
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
}

func NewLoadBalancer(urls []string) *LoadBalancer {
	backends := make([]Backend, len(urls))
	healthChecker := circuit.NewHealthChecker()
	metricsCollector := metrics.NewMetrics()

	for i, u := range urls {
		backends[i] = Backend{
			URL:       u,
			breaker:   circuit.NewCircuitBreaker(),
			isHealthy: true,
		}
		// Start health checking for this backend
		healthChecker.StartChecking(u, time.Second*5) // Check every 5 seconds
	}

	return &LoadBalancer{
		Backends:      backends,
		healthChecker: healthChecker,
		metrics:      metricsCollector,
		client: &http.Client{
			Timeout: time.Second * 2, // 2 second timeout for requests
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
}

// RemoveBackend removes a backend from the load balancer
func (lb *LoadBalancer) RemoveBackend(url string) {
	lb.Lock()
	defer lb.Unlock()

	newBackends := make([]Backend, 0)
	for _, b := range lb.Backends {
		if b.URL != url { // URL is already normalized by admin layer
			newBackends = append(newBackends, b)
		}
	}
	lb.Backends = newBackends
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
	backendEnv := os.Getenv("BACKENDS")
	if backendEnv == "" {
		log.Fatal("BACKENDS env var required")
	}
	urls := strings.Split(backendEnv, ",")
	lb := NewLoadBalancer(urls)

	// Create admin server
	adminServer := admin.NewAdminServer(lb.metrics, lb)

	// Main API endpoint
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		backend := lb.NextBackend()
		if backend == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"No healthy backends available"}`))
			return
		}

		resp, err := lb.forwardRequest(backend, r)
		if err != nil {
			if err == context.DeadlineExceeded {
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte(`{"error":"Backend timeout"}`))
				return
			}
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error":"Backend error"}`))
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// Admin API endpoints
	http.HandleFunc("/admin/metrics", adminServer.HandleMetrics)
	http.HandleFunc("/admin/health", adminServer.HandleHealth)
	http.HandleFunc("/admin/backends", adminServer.HandleBackends)

	log.Println("Round Robin API listening on :8080")
	log.Println("Admin API available at /admin/*")
	http.ListenAndServe(":8080", nil)
}
