package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"round-robin-api/internal/circuit"
)

type Backend struct {
	URL       string
	breaker   *circuit.CircuitBreaker
	isHealthy bool
}

type LoadBalancer struct {
	Backends      []Backend
	currIndex    uint64
	healthChecker *circuit.HealthChecker
	client       *http.Client
}

func NewLoadBalancer(urls []string) *LoadBalancer {
	backends := make([]Backend, len(urls))
	healthChecker := circuit.NewHealthChecker()

	for i, u := range urls {
		backends[i] = Backend{
			URL:     u,
			breaker: circuit.NewCircuitBreaker(),
			isHealthy: true,
		}
		// Start health checking for this backend
		healthChecker.StartChecking(u, time.Second*5) // Check every 5 seconds
	}

	return &LoadBalancer{
		Backends:      backends,
		healthChecker: healthChecker,
		client: &http.Client{
			Timeout: time.Second * 2, // 2 second timeout for requests
		},
	}
}

func (lb *LoadBalancer) NextBackend() *Backend {
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
			return backend
		}
	}
	return nil
}

func (lb *LoadBalancer) forwardRequest(backend *Backend, r *http.Request) (*http.Response, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*2)
	defer cancel()

	// Create new request with timeout context
	req, err := http.NewRequestWithContext(ctx, r.Method, backend.URL+"/", r.Body)
	if err != nil {
		return nil, err
	}
	req.Header = r.Header

	// Forward the request
	resp, err := lb.client.Do(req)
	if err != nil {
		backend.breaker.RecordFailure()
		return nil, err
	}

	// Record success if status code is okay
	if resp.StatusCode < 500 {
		backend.breaker.RecordSuccess()
	} else {
		backend.breaker.RecordFailure()
	}

	return resp, nil
}

func main() {
	backendEnv := os.Getenv("BACKENDS")
	if backendEnv == "" {
		log.Fatal("BACKENDS env var required")
	}
	urls := strings.Split(backendEnv, ",")
	lb := NewLoadBalancer(urls)

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
			// Check if error is a timeout
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

	log.Println("Round Robin API listening on :8080")
	http.ListenAndServe(":8080", nil)
}
