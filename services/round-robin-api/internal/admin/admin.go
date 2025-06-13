package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"round-robin-api/internal/metrics"
)

type AdminServer struct {
	metrics  *metrics.Metrics
	lb       LoadBalancer
}

type LoadBalancer interface {
	AddBackend(url string)
	RemoveBackend(url string)
	GetBackends() []string
}

func NewAdminServer(metrics *metrics.Metrics, lb LoadBalancer) *AdminServer {
	return &AdminServer{
		metrics: metrics,
		lb:      lb,
	}
}

// HandleMetrics returns current metrics
func (s *AdminServer) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	metrics := s.metrics.GetMetrics()
	json.NewEncoder(w).Encode(metrics)
}

// HandleHealth returns backend health status
func (s *AdminServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	health := map[string]interface{}{
		"backends": s.lb.GetBackends(),
		"status":   "ok",
	}
	json.NewEncoder(w).Encode(health)
}

// extractHostPort extracts host:port from a URL
func extractHostPort(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL format")
	}
	return u.Host, nil
}

// validateBackendURL validates and normalizes a backend URL
func validateBackendURL(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %v", err)
	}

	// Ensure scheme is present
	if u.Scheme == "" {
		u.Scheme = "http"
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid URL scheme: must be http or https")
	}

	// Validate host and port
	if u.Host == "" {
		return "", fmt.Errorf("invalid URL: host is required")
	}

	// If port is not specified, use default ports
	if u.Port() == "" {
		if u.Scheme == "https" {
			u.Host = u.Host + ":443"
		} else {
			u.Host = u.Host + ":80"
		}
	}

	return u.String(), nil
}

// HandleBackends manages backend list
func (s *AdminServer) HandleBackends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodGet:
		// List all backends
		backends := s.lb.GetBackends()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"backends": backends,
		})

	case http.MethodPost, http.MethodDelete:
		var backend struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
			http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
			return
		}

		normalizedURL, err := validateBackendURL(backend.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodPost {
			s.lb.AddBackend(normalizedURL)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "added",
				"backend": normalizedURL,
			})
		} else {
			s.lb.RemoveBackend(normalizedURL)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "removed",
				"backend": normalizedURL,
			})
		}

	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
