package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEndpoints(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/health", `{"status":"healthy"}`},
		{"/version", `{"version":"v1.0.0"}`},
		{"/tasks", `{"tasks":[]}`},
	}

	mux := newMux()
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("%s: got status %d, want 200", tt.path, rec.Code)
		}
		if got := strings.TrimSpace(rec.Body.String()); got != tt.want {
			t.Errorf("%s: got body %s, want %s", tt.path, got, tt.want)
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/health", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("got status %d, want 405", rec.Code)
	}
}
