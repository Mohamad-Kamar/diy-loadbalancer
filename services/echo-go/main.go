package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[go] Received request: %s %s\n", r.Method, r.URL.Path)
	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
	fmt.Println("[go] Request processed successfully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	http.HandleFunc("/", echoHandler)
	http.HandleFunc("/health", healthHandler)
	fmt.Printf("[go] Server started on port %d\n", 8081)
	http.ListenAndServe(":8081", nil)
}
