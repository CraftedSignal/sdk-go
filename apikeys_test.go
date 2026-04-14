package craftedsignal_test

import (
	"context"
	"net/http"
	"testing"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

func TestAPIKeysList(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/api-keys" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, []map[string]any{
			{"id": "key-1", "name": "ci-key", "key_prefix": "cskey_ab"},
		})
	}))
	defer cleanup()

	keys, err := client.APIKeys.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0].ID != "key-1" {
		t.Errorf("unexpected keys: %+v", keys)
	}
}

func TestAPIKeysCreate(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/api-keys" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"id": "key-2", "name": "new-key", "key_prefix": "cskey_cd",
			"key": "cskey_cdXXXXXXXXXXXXXXXXXXXXXX",
		})
	}))
	defer cleanup()

	k, err := client.APIKeys.Create(context.Background(), craftedsignal.CreateAPIKeyRequest{
		Name:   "new-key",
		Scopes: []string{"rules:read"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if k.PlaintextKey == "" {
		t.Error("PlaintextKey should be non-empty on creation")
	}
}

func TestAPIKeysRevoke(t *testing.T) {
	var gotMethod, gotPath string
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer cleanup()

	err := client.APIKeys.Revoke(context.Background(), "key-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v1/api-keys/key-1" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}
}
