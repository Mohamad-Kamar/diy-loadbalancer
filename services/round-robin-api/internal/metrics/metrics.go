package metrics

import (
	"sync"
	"time"

	"round-robin-api/internal/circuit"
)

// Metrics stores various load balancer metrics
type Metrics struct {
	mu            sync.RWMutex
	RequestCounts map[string]uint64
	ResponseTimes map[string]time.Duration
	ErrorRates    map[string]float64
	CircuitStates map[string]circuit.State
	totalErrors   map[string]uint64
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		RequestCounts:  make(map[string]uint64),
		ResponseTimes: make(map[string]time.Duration),
		ErrorRates:    make(map[string]float64),
		CircuitStates: make(map[string]circuit.State),
		totalErrors:   make(map[string]uint64),
	}
}

// RecordRequest records a request to a backend
func (m *Metrics) RecordRequest(backend string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RequestCounts[backend]++
}

// RecordResponseTime records the response time for a backend request
func (m *Metrics) RecordResponseTime(backend string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Use exponential moving average for response time
	if curr, exists := m.ResponseTimes[backend]; exists {
		m.ResponseTimes[backend] = (curr*9 + duration) / 10
	} else {
		m.ResponseTimes[backend] = duration
	}
}

// RecordError records an error for a backend
func (m *Metrics) RecordError(backend string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalErrors[backend]++
	if total := m.RequestCounts[backend]; total > 0 {
		m.ErrorRates[backend] = float64(m.totalErrors[backend]) / float64(total)
	}
}

// RecordCircuitState records the current circuit state for a backend
func (m *Metrics) RecordCircuitState(backend string, state circuit.State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CircuitStates[backend] = state
}

// GetMetrics returns a copy of current metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"request_counts":  m.RequestCounts,
		"response_times": m.ResponseTimes,
		"error_rates":    m.ErrorRates,
		"circuit_states": m.CircuitStates,
	}
}
