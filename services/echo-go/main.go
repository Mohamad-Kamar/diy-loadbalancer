package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Echo go request: POST /")
	var payload map[string]interface{}
	json.NewDecoder(r.Body).Decode(&payload)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
	log.Println("Echo go response sent successfully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	http.HandleFunc("/", echoHandler)
	http.HandleFunc("/health", healthHandler)
	log.Println("Go echo server started on port 8081")
	http.ListenAndServe(":8081", nil)
}
