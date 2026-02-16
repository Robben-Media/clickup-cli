package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet_SetsAuthHeaderAndDecodesResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "test-api-key" {
			t.Fatalf("expected Authorization test-api-key, got %s", r.Header.Get("Authorization"))
		}

		if r.URL.Path != "/tasks/1" {
			t.Fatalf("expected path /tasks/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   "task-1",
			"name": "Test task",
		})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := client.Get(context.Background(), "/tasks/1", &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "task-1" {
		t.Fatalf("expected id task-1, got %s", result.ID)
	}
}

func TestPost_SendsJSONBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		var req map[string]string
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req["name"] != "Task from test" {
			t.Fatalf("expected request name Task from test, got %s", req["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result struct {
		OK bool `json:"ok"`
	}

	if err := client.Post(context.Background(), "/tasks", map[string]string{"name": "Task from test"}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.OK {
		t.Fatal("expected ok=true")
	}
}

func TestDelete_PropagatesAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	err := client.Delete(context.Background(), "/tasks/1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", apiErr.StatusCode)
	}

	if apiErr.Message != "forbidden" {
		t.Fatalf("expected message forbidden, got %s", apiErr.Message)
	}
}

func TestGet_HTTPErrorFallbackStatusText(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream bad gateway body"))
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result map[string]any
	err := client.Get(context.Background(), "/tasks/1", &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.Message != "Bad Gateway" {
		t.Fatalf("expected fallback message Bad Gateway, got %s", apiErr.Message)
	}
}
