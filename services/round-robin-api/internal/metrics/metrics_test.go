package metrics

import (
	"fmt"
	"round-robin-api/internal/circuit"
	"testing"
	"time"
)

func TestMetrics_NewMetrics(t *testing.T) {
	m := NewMetrics()
	if m.RequestCounts == nil {
		t.Error("RequestCounts map should be initialized")
	}
	if m.ResponseTimes == nil {
		t.Error("ResponseTimes map should be initialized")
	}
	if m.ErrorRates == nil {
		t.Error("ErrorRates map should be initialized")
	}
	if m.CircuitStates == nil {
		t.Error("CircuitStates map should be initialized")
	}
	if m.maxRecents != 100 {
		t.Error("maxRecents should be initialized to 100")
	}
}

func TestMetrics_RecordRequest(t *testing.T) {
	m := NewMetrics()
	backend := "test-backend"
	
	// Test single request
	m.RecordRequest(backend)
	if count := m.RequestCounts[backend]; count != 1 {
		t.Errorf("Expected request count 1, got %d", count)
	}
	
	// Test multiple requests
	m.RecordRequest(backend)
	m.RecordRequest(backend)
	if count := m.RequestCounts[backend]; count != 3 {
		t.Errorf("Expected request count 3, got %d", count)
	}
}

func TestMetrics_RecordResponseTime(t *testing.T) {
	m := NewMetrics()
	backend := "test-backend"
	duration := time.Duration(100 * time.Millisecond)
	
	// Test first response time
	m.RecordResponseTime(backend, duration)
	if rt := m.ResponseTimes[backend]; rt != duration {
		t.Errorf("Expected response time %v, got %v", duration, rt)
	}
	
	// Test exponential moving average
	secondDuration := time.Duration(200 * time.Millisecond)
	m.RecordResponseTime(backend, secondDuration)
	expected := (duration*9 + secondDuration) / 10
	if rt := m.ResponseTimes[backend]; rt != expected {
		t.Errorf("Expected response time %v, got %v", expected, rt)
	}
}

func TestMetrics_RecordError(t *testing.T) {
	m := NewMetrics()
	backend := "test-backend"
	
	// Record some requests first
	m.RequestCounts[backend] = 10
	
	// Test error recording
	m.RecordError(backend)
	if rate := m.ErrorRates[backend]; rate != 0.1 {
		t.Errorf("Expected error rate 0.1, got %f", rate)
	}
	
	// Test multiple errors
	m.RecordError(backend)
	if rate := m.ErrorRates[backend]; rate != 0.2 {
		t.Errorf("Expected error rate 0.2, got %f", rate)
	}
}

func TestMetrics_RecordCircuitState(t *testing.T) {
	m := NewMetrics()
	backend := "test-backend"
	
	states := []circuit.State{
		circuit.CLOSED,
		circuit.OPEN,
		circuit.HALF_OPEN,
	}
	
	for _, state := range states {
		m.RecordCircuitState(backend, state)
		if s := m.CircuitStates[backend]; s != state {
			t.Errorf("Expected circuit state %v, got %v", state, s)
		}
	}
}

func TestMetrics_RecordRequestComplete(t *testing.T) {
	m := NewMetrics()
	
	tests := []struct {
		name     string
		id       string
		backend  string
		duration time.Duration
		success  bool
	}{
		{
			name:     "successful request",
			id:       "req1",
			backend:  "backend1",
			duration: 100 * time.Millisecond,
			success:  true,
		},
		{
			name:     "failed request",
			id:       "req2",
			backend:  "backend1",
			duration: 200 * time.Millisecond,
			success:  false,
		},
		{
			name:     "different backend",
			id:       "req3",
			backend:  "backend2",
			duration: 150 * time.Millisecond,
			success:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.RecordRequestComplete(tt.id, tt.backend, tt.duration, tt.success)
			
			// Check metrics
			metrics := m.GetMetrics()
			
			// Verify request was counted
			if count := metrics["request_counts"].(map[string]uint64)[tt.backend]; count != 1 {
				t.Errorf("Expected request count 1, got %d", count)
			}
			
			// Verify response time was recorded
			if _, exists := metrics["response_times"].(map[string]time.Duration)[tt.backend]; !exists {
				t.Errorf("Response time was not recorded for %s", tt.backend)
			}
			
			// Verify error rate for failed requests
			if !tt.success {
				if rate := metrics["error_rates"].(map[string]float64)[tt.backend]; rate != 1.0 {
					t.Errorf("Expected error rate 1.0 for failed request, got %f", rate)
				}
			}
			
			// Verify recent requests
			recents := metrics["recent_requests"].([]RequestInfo)
			if len(recents) == 0 {
				t.Error("No recent requests recorded")
			} else {
				recent := recents[0]
				if recent.ID != tt.id {
					t.Errorf("Expected recent request ID %s, got %s", tt.id, recent.ID)
				}
				if recent.Backend != tt.backend {
					t.Errorf("Expected recent request backend %s, got %s", tt.backend, recent.Backend)
				}
				if recent.Duration != tt.duration {
					t.Errorf("Expected recent request duration %v, got %v", tt.duration, recent.Duration)
				}
				if recent.Success != tt.success {
					t.Errorf("Expected recent request success %v, got %v", tt.success, recent.Success)
				}
			}
		})
	}
}

func TestMetrics_RecentRequestsLimit(t *testing.T) {
	m := NewMetrics()
	backend := "test-backend"
	
	// Add more requests than the limit
	for i := 0; i < m.maxRecents+10; i++ {
		m.RecordRequestComplete(
			fmt.Sprintf("req%d", i),
			backend,
			100*time.Millisecond,
			true,
		)
	}
	
	metrics := m.GetMetrics()
	recents := metrics["recent_requests"].([]RequestInfo)
	
	if len(recents) > m.maxRecents {
		t.Errorf("Recent requests exceeded limit, got %d want <= %d",
			len(recents), m.maxRecents)
	}
	
	// Verify order (most recent first)
	for i := 0; i < len(recents)-1; i++ {
		if recents[i].Timestamp.Before(recents[i+1].Timestamp) {
			t.Error("Recent requests not in correct order")
			break
		}
	}
}
