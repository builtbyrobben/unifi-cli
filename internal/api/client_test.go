package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T) (*httptest.Server, *int) {
	t.Helper()

	loginCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("login: expected POST, got %s", r.Method)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)

		if body["username"] != "admin" || body["password"] != "secret" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})

			return
		}

		loginCount++

		http.SetCookie(w, &http.Cookie{
			Name:  "TOKEN",
			Value: "test-session-token",
			Path:  "/",
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/device", func(w http.ResponseWriter, r *http.Request) {
		// Check for session cookie
		cookie, err := r.Cookie("TOKEN")
		if err != nil || cookie.Value != "test-session-token" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"mac": "AA:BB:CC:DD:EE:FF", "name": "Switch", "type": "usw"},
			},
			"meta": map[string]string{"rc": "ok"},
		})
	})

	server := httptest.NewServer(mux)

	return server, &loginCount
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	client, err := NewClient(Credentials{
		Host:     server.URL,
		Username: "admin",
		Password: "secret",
		Site:     "default",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	return client
}

func TestNewClient_Validation(t *testing.T) {
	tests := []struct {
		name    string
		creds   Credentials
		wantErr bool
	}{
		{
			name:    "valid credentials",
			creds:   Credentials{Host: "https://localhost", Username: "admin", Password: "pass"},
			wantErr: false,
		},
		{
			name:    "missing host",
			creds:   Credentials{Username: "admin", Password: "pass"},
			wantErr: true,
		},
		{
			name:    "missing username",
			creds:   Credentials{Host: "https://localhost", Password: "pass"},
			wantErr: true,
		},
		{
			name:    "missing password",
			creds:   Credentials{Host: "https://localhost", Username: "admin"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.creds)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewClient_DefaultSite(t *testing.T) {
	client, err := NewClient(Credentials{
		Host:     "https://localhost",
		Username: "admin",
		Password: "pass",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.SitePath() != "/proxy/network/api/s/default" {
		t.Errorf("expected default site path, got %q", client.SitePath())
	}
}

func TestClient_Login(t *testing.T) {
	server, loginCount := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Login(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *loginCount != 1 {
		t.Errorf("expected 1 login, got %d", *loginCount)
	}
}

func TestClient_Login_InvalidCredentials(t *testing.T) {
	server, _ := newTestServer(t)
	defer server.Close()

	client, err := NewClient(Credentials{
		Host:     server.URL,
		Username: "admin",
		Password: "wrong",
		Site:     "default",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	err = client.Login(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
}

func TestClient_AutoLogin(t *testing.T) {
	server, loginCount := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	// First request should trigger auto-login
	var result map[string]any

	err := client.Get(context.Background(), "/proxy/network/api/s/default/stat/device", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *loginCount != 1 {
		t.Errorf("expected 1 auto-login, got %d", *loginCount)
	}
}

func TestClient_AutoReLogin_On401(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			http.SetCookie(w, &http.Cookie{
				Name:  "TOKEN",
				Value: "token",
				Path:  "/",
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

			return
		}

		callCount++

		// Return 401 on first call, 200 on retry
		if callCount == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []any{},
			"meta": map[string]string{"rc": "ok"},
		})
	}))
	defer server.Close()

	client, err := NewClient(Credentials{
		Host:     server.URL,
		Username: "admin",
		Password: "secret",
		Site:     "default",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	// Pre-authenticate
	client.authenticated = true

	var result map[string]any

	err = client.Get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls (initial 401 + retry), got %d", callCount)
	}
}

func TestClient_Get(t *testing.T) {
	server, _ := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	var result map[string]any

	err := client.Get(context.Background(), "/proxy/network/api/s/default/stat/device", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result["data"].([]any)
	if !ok {
		t.Fatal("expected data array in response")
	}

	if len(data) != 1 {
		t.Fatalf("expected 1 device, got %d", len(data))
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			http.SetCookie(w, &http.Cookie{Name: "TOKEN", Value: "token", Path: "/"})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

			return
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": []any{},
			"meta": map[string]string{"rc": "ok"},
		})
	}))
	defer server.Close()

	client, err := NewClient(Credentials{
		Host:     server.URL,
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	var result map[string]any
	body := map[string]string{"cmd": "restart", "mac": "AA:BB:CC:DD:EE:FF"}

	err = client.Post(context.Background(), "/cmd/devmgr", body, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/login" {
			http.SetCookie(w, &http.Cookie{Name: "TOKEN", Value: "token", Path: "/"})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]string{"rc": "error", "msg": "api.err.UnknownDevice"},
		})
	}))
	defer server.Close()

	client, err := NewClient(Credentials{
		Host:     server.URL,
		Username: "admin",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	var result map[string]any

	err = client.Get(context.Background(), "/test", &result)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}

	if apiErr.Message != "api.err.UnknownDevice" {
		t.Errorf("expected message 'api.err.UnknownDevice', got %q", apiErr.Message)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 404, Message: "Not Found"}
	expected := "API error (404): Not Found"

	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestSitePath(t *testing.T) {
	client, err := NewClient(Credentials{
		Host:     "https://localhost",
		Username: "admin",
		Password: "pass",
		Site:     "mysite",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "/proxy/network/api/s/mysite"
	if client.SitePath() != expected {
		t.Errorf("SitePath() = %q, want %q", client.SitePath(), expected)
	}
}

func TestWithUserAgent(t *testing.T) {
	client, err := NewClient(
		Credentials{Host: "https://localhost", Username: "admin", Password: "pass"},
		WithUserAgent("test-agent/2.0"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.userAgent != "test-agent/2.0" {
		t.Errorf("expected userAgent 'test-agent/2.0', got %q", client.userAgent)
	}
}

func TestWithTimeout(t *testing.T) {
	client, err := NewClient(
		Credentials{Host: "https://localhost", Username: "admin", Password: "pass"},
		WithTimeout(5000000000),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.httpClient.Timeout != 5000000000 {
		t.Errorf("expected timeout 5s, got %v", client.httpClient.Timeout)
	}
}
