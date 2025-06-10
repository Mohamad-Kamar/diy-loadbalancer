package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type Backend struct {
	URL string
}

type LoadBalancer struct {
	Backends  []Backend
	currIndex uint64
}

func (lb *LoadBalancer) NextBackend() *Backend {
	if len(lb.Backends) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&lb.currIndex, 1) % uint64(len(lb.Backends))
	return &lb.Backends[idx]
}

func main() {
	backendEnv := os.Getenv("BACKENDS")
	if backendEnv == "" {
		log.Fatal("BACKENDS env var required")
	}
	urls := strings.Split(backendEnv, ",")
	backends := make([]Backend, len(urls))
	for i, u := range urls {
		backends[i] = Backend{URL: u}
	}
	lb := &LoadBalancer{Backends: backends}

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		backend := lb.NextBackend()
		if backend == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"No backends available"}`))
			return
		}
		resp, err := http.Post(backend.URL+"/", "application/json", r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error":"Backend error"}`))
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	log.Println("Round Robin API listening on :8080")
	http.ListenAndServe(":8080", nil)
}
