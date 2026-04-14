package craftedsignal

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestTransport(srv *httptest.Server) *transport {
	return &transport{
		token:        Token("test-token"),
		baseURL:      srv.URL,
		httpClient:   srv.Client(),
		maxRetries:   0,
		backoff:      NoRetry,
		logger:       discardLogger(),
		pollInterval: 10 * time.Millisecond,
	}
}

func writeTestAPIResponse(w http.ResponseWriter, status int, jsonData string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"success":true,"data":` + jsonData + `}`))
}

func TestTransportAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeTestAPIResponse(w, 200, `{}`)
	}))
	defer srv.Close()

	tr := newTestTransport(srv)
	resp, err := tr.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-token")
	}
}

func TestTransportRetry429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		writeTestAPIResponse(w, 200, `{"ok":true}`)
	}))
	defer srv.Close()

	tr := newTestTransport(srv)
	tr.maxRetries = 3
	tr.backoff = func(_ int) time.Duration { return time.Millisecond }

	resp, err := tr.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestTransportContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		writeTestAPIResponse(w, 200, `null`)
	}))
	defer srv.Close()

	tr := newTestTransport(srv)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := tr.do(ctx, "GET", "/test", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestTransportDecode401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"unauthorized","message":"bad token"}}`))
	}))
	defer srv.Close()

	tr := newTestTransport(srv)
	resp, err := tr.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	err = tr.decode(resp, &out)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestTransportDecode404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"success":false,"error":{"code":"not_found","message":"rule not found"}}`))
	}))
	defer srv.Close()

	tr := newTestTransport(srv)
	resp, err := tr.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = tr.decode(resp, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTransportRetry5xx(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(500)
			return
		}
		writeTestAPIResponse(w, 200, `null`)
	}))
	defer srv.Close()

	tr := newTestTransport(srv)
	tr.maxRetries = 2
	tr.backoff = func(_ int) time.Duration { return time.Millisecond }

	resp, err := tr.do(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}
