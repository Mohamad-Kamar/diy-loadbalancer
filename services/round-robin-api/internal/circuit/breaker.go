package circuit

import (
	"sync"
	"time"
)

// State represents the circuit breaker states
type State int

const (
	CLOSED State = iota
	OPEN
	HALF_OPEN
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	sync.RWMutex
	state            State
	failureThreshold int
	failureCount     int
	lastFailureTime  time.Time
	timeout          time.Duration
}

// NewCircuitBreaker creates a new circuit breaker with default settings
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		state:            CLOSED,
		failureThreshold: 5,            // After 5 failures, circuit opens
		timeout:          time.Second * 10, // Wait 10 seconds before trying again
	}
}

// IsAvailable checks if the circuit is closed or can be tested (half-open)
func (cb *CircuitBreaker) IsAvailable() bool {
	cb.RLock()
	defer cb.RUnlock()

	switch cb.state {
	case CLOSED:
		return true
	case OPEN:
		// Check if enough time has passed to move to half-open
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.toHalfOpen()
			return true
		}
		return false
	case HALF_OPEN:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request and closes the circuit if in half-open state
func (cb *CircuitBreaker) RecordSuccess() {
	cb.Lock()
	defer cb.Unlock()

	if cb.state == HALF_OPEN {
		cb.state = CLOSED
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request and may open the circuit
func (cb *CircuitBreaker) RecordFailure() {
	cb.Lock()
	defer cb.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == CLOSED && cb.failureCount >= cb.failureThreshold {
		cb.state = OPEN
	} else if cb.state == HALF_OPEN {
		cb.state = OPEN
	}
}

// toHalfOpen moves the circuit to half-open state (helper function)
func (cb *CircuitBreaker) toHalfOpen() {
	cb.Lock()
	defer cb.Unlock()

	if cb.state == OPEN {
		cb.state = HALF_OPEN
		cb.failureCount = 0
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.RLock()
	defer cb.RUnlock()
	return cb.state
}
