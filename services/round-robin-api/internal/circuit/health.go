package circuit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthChecker manages health checking of backend servers
type HealthChecker struct {
	sync.RWMutex
	healthStatus map[string]bool
	client       *http.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		healthStatus: make(map[string]bool),
		client: &http.Client{
			Timeout: time.Second * 2, // 2 second timeout for health checks
		},
	}
}

// StartChecking begins periodic health checks for a backend
func (hc *HealthChecker) StartChecking(url string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			<-ticker.C
			isHealthy := hc.checkHealth(url)
			hc.setHealth(url, isHealthy)
		}
	}()
}

// IsHealthy returns whether a backend is currently healthy
func (hc *HealthChecker) IsHealthy(url string) bool {
	hc.RLock()
	defer hc.RUnlock()
	return hc.healthStatus[url]
}

// checkHealth performs a single health check
func (hc *HealthChecker) checkHealth(url string) bool {
	healthURL := fmt.Sprintf("%s/health", url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return false
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result["status"] == "ok"
}

// setHealth updates the health status of a backend
func (hc *HealthChecker) setHealth(url string, isHealthy bool) {
	hc.Lock()
	defer hc.Unlock()
	hc.healthStatus[url] = isHealthy
}
