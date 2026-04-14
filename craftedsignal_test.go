package craftedsignal_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

// newTestClient creates a Client wired to an httptest.Server.
// Use this helper in all _test.go files in package craftedsignal_test.
func newTestClient(t *testing.T, handler http.Handler) (*craftedsignal.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client, err := craftedsignal.NewClient("test-token",
		craftedsignal.WithBaseURL(srv.URL),
		craftedsignal.WithRetry(0, craftedsignal.NoRetry),
	)
	if err != nil {
		t.Fatal(err)
	}
	return client, srv.Close
}

// writeJSON writes a standard API envelope response.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]any{"success": true, "data": data}
	_ = json.NewEncoder(w).Encode(resp)
}

func TestNewClientEmptyToken(t *testing.T) {
	_, err := craftedsignal.NewClient("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestMe(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/me" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"company":      "Acme Corp",
			"api_key_name": "ci-key",
			"scopes":       []string{"rules:read"},
		})
	}))
	defer cleanup()

	me, err := client.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if me.Company != "Acme Corp" {
		t.Errorf("Company = %q, want %q", me.Company, "Acme Corp")
	}
	if me.APIKeyName != "ci-key" {
		t.Errorf("APIKeyName = %q, want ci-key", me.APIKeyName)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer cleanup()

	_, err := client.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		writeJSON(w, 200, map[string]any{"company": "Test", "api_key_name": "k", "scopes": []string{}})
	}))
	defer srv.Close()

	client, err := craftedsignal.NewClient("test-token",
		craftedsignal.WithBaseURL(srv.URL),
		craftedsignal.WithRetry(0, craftedsignal.NoRetry),
		craftedsignal.WithUserAgent("myapp/1.0"),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = client.Me(context.Background())
	if gotUA != "myapp/1.0" {
		t.Errorf("User-Agent = %q, want myapp/1.0", gotUA)
	}
}

func TestWithInsecure(t *testing.T) {
	// WithInsecure should not panic and client should be usable
	client, err := craftedsignal.NewClient("test-token",
		craftedsignal.WithInsecure(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestWithVerbose(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"company": "T", "api_key_name": "k", "scopes": []string{}})
	}))
	defer srv.Close()

	client, err := craftedsignal.NewClient("test-token",
		craftedsignal.WithBaseURL(srv.URL),
		craftedsignal.WithRetry(0, craftedsignal.NoRetry),
		craftedsignal.WithVerbose(),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithPollInterval(t *testing.T) {
	// WithPollInterval should set poll interval without panicking
	client, err := craftedsignal.NewClient("test-token",
		craftedsignal.WithPollInterval(10*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}
