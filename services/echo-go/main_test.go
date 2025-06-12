package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEchoHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    string
		expectedStatus int
		expectedBody   string
		contentType    string
	}{
		{
			name:           "valid JSON POST request",
			method:         "POST",
			requestBody:    `{"msg":"hi"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"msg":"hi"}`,
			contentType:    "application/json",
		},
		{
			name:           "invalid method GET",
			method:         "GET",
			expectedStatus: http.StatusMethodNotAllowed,
			contentType:    "application/json",
		},
		{
			name:           "empty POST body",
			method:         "POST",
			requestBody:    "",
			expectedStatus: http.StatusOK,
			expectedBody:   "null",
			contentType:    "application/json",
		},
		{
			name:           "complex JSON object",
			method:         "POST",
			requestBody:    `{"msg":"hello","nested":{"key":"value"},"array":[1,2,3]}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"msg":"hello","nested":{"key":"value"},"array":[1,2,3]}`,
			contentType:    "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			echoHandler(w, req)
			resp := w.Result()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				var expected, got interface{}
				json.Unmarshal([]byte(tt.expectedBody), &expected)
				json.Unmarshal(body, &got)
				if !jsonEqual(expected, got) {
					t.Errorf("expected body %s, got %s", tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok"}`,
		},
		{
			name:           "POST request (not allowed)",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			healthHandler(w, req)
			resp := w.Result()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if strings.TrimSpace(string(body)) != strings.TrimSpace(tt.expectedBody) {
					t.Errorf("expected body %s, got %s", tt.expectedBody, string(body))
				}
			}
		})
	}
}

// Helper function to compare JSON values
func jsonEqual(a, b interface{}) bool {
	switch v := a.(type) {
	case map[string]interface{}:
		bMap, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(v) != len(bMap) {
			return false
		}
		for k, av := range v {
			bv, ok := bMap[k]
			if !ok {
				return false
			}
			if !jsonEqual(av, bv) {
				return false
			}
		}
		return true
	case []interface{}:
		bArray, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(v) != len(bArray) {
			return false
		}
		for i, av := range v {
			if !jsonEqual(av, bArray[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
