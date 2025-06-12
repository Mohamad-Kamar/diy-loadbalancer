package circuit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthChecker_NewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	if hc == nil {
		t.Fatal("HealthChecker should not be nil")
	}
	if hc.healthStatus == nil {
		t.Error("healthStatus map should be initialized")
	}
	if hc.client == nil {
		t.Error("client should be initialized")
	}
	if hc.client.Timeout != time.Second*2 {
		t.Error("client timeout should be 2 seconds")
	}
}

func TestHealthChecker_IsHealthy(t *testing.T) {
	hc := NewHealthChecker()
	backend := "http://test-backend"
	
	// Test unknown backend (should be unhealthy)
	if hc.IsHealthy(backend) {
		t.Error("Unknown backend should be considered unhealthy")
	}
	
	// Test setting health status
	hc.setHealth(backend, true)
	if !hc.IsHealthy(backend) {
		t.Error("Backend should be healthy after setting health to true")
	}
	
	hc.setHealth(backend, false)
	if hc.IsHealthy(backend) {
		t.Error("Backend should be unhealthy after setting health to false")
	}
}

func TestHealthChecker_CheckHealth(t *testing.T) {
	tests := []struct {
		name           string
		handlerFunc    http.HandlerFunc
		expectedHealth bool
	}{
		{
			name: "healthy backend",
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			}),
			expectedHealth: true,
		},
		{
			name: "unhealthy backend - wrong status",
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{"status": "error"})
			}),
			expectedHealth: false,
		},
		{
			name: "unhealthy backend - invalid response",
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not json"))
			}),
			expectedHealth: false,
		},
		{
			name: "unhealthy backend - wrong status format",
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "not ok"})
			}),
			expectedHealth: false,
		},
		{
			name: "unhealthy backend - timeout",
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(3 * time.Second)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			}),
			expectedHealth: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handlerFunc)
			defer server.Close()
			
			hc := NewHealthChecker()
			isHealthy := hc.checkHealth(server.URL)
			
			if isHealthy != tt.expectedHealth {
				t.Errorf("Expected health check %v, got %v", tt.expectedHealth, isHealthy)
			}
		})
	}
}

func TestHealthChecker_StartChecking(t *testing.T) {
	healthChanges := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case healthy := <-healthChanges:
			if healthy {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		default:
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}
	}))
	defer server.Close()
	
	hc := NewHealthChecker()
	interval := 100 * time.Millisecond
	
	// Start health checking
	hc.StartChecking(server.URL, interval)
	
	// Wait for initial health check
	time.Sleep(interval * 2)
	
	// Verify initial health status
	if !hc.IsHealthy(server.URL) {
		t.Error("Backend should be initially healthy")
	}
	
	// Make backend unhealthy
	healthChanges <- false
	
	// Wait for next health check
	time.Sleep(interval * 2)
	
	// Verify backend is now unhealthy
	if hc.IsHealthy(server.URL) {
		t.Error("Backend should be unhealthy after failure")
	}
	
	// Make backend healthy again
	healthChanges <- true
	
	// Wait for next health check
	time.Sleep(interval * 2)
	
	// Verify backend is healthy again
	if !hc.IsHealthy(server.URL) {
		t.Error("Backend should be healthy after recovery")
	}
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	hc := NewHealthChecker()
	backend := "http://test-backend"
	
	// Simulate concurrent health status updates
	go func() {
		for i := 0; i < 100; i++ {
			hc.setHealth(backend, i%2 == 0)
		}
	}()
	
	// Simulate concurrent health checks
	go func() {
		for i := 0; i < 100; i++ {
			_ = hc.IsHealthy(backend)
		}
	}()
	
	// Wait for goroutines to finish
	time.Sleep(100 * time.Millisecond)
}
