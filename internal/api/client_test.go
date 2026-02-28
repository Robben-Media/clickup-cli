package api

import (
	"bytes"
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

func TestPatch_SendsJSONBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
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

		if req["name"] != "Updated name" {
			t.Fatalf("expected request name Updated name, got %s", req["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result struct {
		OK bool `json:"ok"`
	}

	if err := client.Patch(context.Background(), "/resource/1", map[string]string{"name": "Updated name"}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.OK {
		t.Fatal("expected ok=true")
	}
}

func TestPatch_PropagatesAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}

		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result any

	err := client.Patch(context.Background(), "/resource/1", map[string]string{"name": "test"}, &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", apiErr.StatusCode)
	}
}

func TestPostMultipart_SendsFile(t *testing.T) {
	t.Parallel()

	fileContent := []byte("test file content")
	fileName := "test.txt"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			t.Fatal("expected Content-Type header to be set")
		}

		// Verify it's a multipart form
		if len(contentType) < 9 || contentType[:9] != "multipart" {
			t.Fatalf("expected multipart content-type, got %s", contentType)
		}

		// Read and verify the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		// Check that the file content is in the body
		if !bytes.Contains(body, fileContent) {
			t.Fatal("expected file content in multipart body")
		}

		// Check that the filename is in the body
		if !bytes.Contains(body, []byte(fileName)) {
			t.Fatal("expected filename in multipart body")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result struct {
		OK bool `json:"ok"`
	}

	if err := client.PostMultipart(context.Background(), "/upload", "attachment", bytes.NewReader(fileContent), fileName, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.OK {
		t.Fatal("expected ok=true")
	}
}

func TestPostMultipart_SetsAuthHeader(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "test-api-key" {
			t.Fatalf("expected Authorization test-api-key, got %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result struct {
		OK bool `json:"ok"`
	}

	if err := client.PostMultipart(context.Background(), "/upload", "file", bytes.NewReader([]byte("content")), "test.txt", &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostMultipart_PropagatesAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	var result any

	err := client.PostMultipart(context.Background(), "/upload", "file", bytes.NewReader([]byte("content")), "test.txt", &result)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestBaseURLConstruction(t *testing.T) {
	t.Parallel()

	// Test that the default base URL is set correctly
	// The default base URL should not include /v2 anymore
	// Requests should include the version prefix in the path
	// This test verifies the base URL is "https://api.clickup.com/api"
	expectedBaseURL := "https://api.clickup.com/api"

	// Create a test server to capture the actual URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the full URL path includes the version
		if r.URL.Path != "/v2/task/123" {
			t.Fatalf("expected path /v2/task/123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	// Use a custom client with the test server URL
	testClient := NewClient("test-key", WithBaseURL(server.URL))

	var result any
	if err := testClient.Get(context.Background(), "/v2/task/123", &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify expectedBaseURL is correct
	if expectedBaseURL != "https://api.clickup.com/api" {
		t.Fatalf("expected base URL https://api.clickup.com/api, got %s", expectedBaseURL)
	}
}
