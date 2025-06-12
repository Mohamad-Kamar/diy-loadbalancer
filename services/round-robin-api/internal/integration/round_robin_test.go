package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestRoundRobinDistribution(t *testing.T) {
	baseURL := "http://localhost:8080"
	
	// Test backend management
	t.Run("Backend Management", func(t *testing.T) {
		// Get initial backends
		resp, err := http.Get(baseURL + "/admin/backends")
		if err != nil {
			t.Fatalf("Failed to get backends: %v", err)
		}
		var backends []string
		json.NewDecoder(resp.Body).Decode(&backends)
		initialCount := len(backends)

		// Add new backend
		newBackend := "http://localhost:8084"
		addReq := struct {
			URL string `json:"url"`
		}{URL: newBackend}
		body, _ := json.Marshal(addReq)
		
		resp, err = http.Post(baseURL+"/admin/backends", "application/json", bytes.NewBuffer(body))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to add backend: %v", err)
		}

		// Verify backend was added
		resp, _ = http.Get(baseURL + "/admin/backends")
		json.NewDecoder(resp.Body).Decode(&backends)
		if len(backends) != initialCount+1 {
			t.Errorf("Backend count should be %d, got %d", initialCount+1, len(backends))
		}

		// Remove backend
		req, _ := http.NewRequest(http.MethodDelete, baseURL+"/admin/backends", bytes.NewBuffer(body))
		resp, err = http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to remove backend: %v", err)
		}

		// Verify backend was removed
		resp, _ = http.Get(baseURL + "/admin/backends")
		json.NewDecoder(resp.Body).Decode(&backends)
		if len(backends) != initialCount {
			t.Errorf("Backend count should be %d, got %d", initialCount, len(backends))
		}
	})

	// Test round robin distribution
	t.Run("Request Distribution", func(t *testing.T) {
		requestCount := 30
		results := make(map[string]int)

		// Send requests
		for i := 0; i < requestCount; i++ {
			payload := map[string]interface{}{
				"test": i,
			}
			body, _ := json.Marshal(payload)
			resp, err := http.Post(baseURL+"/api", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			// Extract backend from response headers
			backend := resp.Header.Get("X-Served-by")
			results[backend]++
		}

		// Verify even distribution (with some tolerance)
		expectedPerBackend := requestCount / 3
		tolerance := 2

		for backend, count := range results {
			if count < expectedPerBackend-tolerance || count > expectedPerBackend+tolerance {
				t.Errorf("Uneven distribution for %s: got %d requests, expected ~%d (±%d)", 
					backend, count, expectedPerBackend, tolerance)
			}
		}
	})

	// Test fault tolerance
	t.Run("Fault Tolerance", func(t *testing.T) {
		// Get metrics before test
		resp, _ := http.Get(baseURL + "/admin/metrics")
		var beforeMetrics map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&beforeMetrics)

		// Send requests with potential failures
		for i := 0; i < 10; i++ {
			payload := map[string]interface{}{
				"test": i,
			}
			body, _ := json.Marshal(payload)
			resp, err := http.Post(baseURL+"/api", "application/json", bytes.NewBuffer(body))
			if err != nil {
				continue // Expected some failures
			}
			defer resp.Body.Close()
		}

		// Wait for circuit breaker recovery
		time.Sleep(5 * time.Second)

		// Get metrics after test
		resp, _ = http.Get(baseURL + "/admin/metrics")
		var afterMetrics map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&afterMetrics)

		// Verify error rates and circuit breaker state
		// The actual validation would depend on your metrics structure
		if afterMetrics["total_requests"].(float64) <= beforeMetrics["total_requests"].(float64) {
			t.Error("No requests were processed during fault tolerance test")
		}
	})
}
