package http

import (
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterServesPing(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(nethttp.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["message"] != "pong" {
		t.Fatalf("expected message pong, got %q", body["message"])
	}
}
