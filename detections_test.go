package craftedsignal_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

func TestDetectionsExport(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/detections/export" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, []map[string]any{
			{"id": "rule-1", "title": "Test Rule", "platform": "splunk", "enabled": true},
		})
	}))
	defer cleanup()

	rules, err := client.Detections.Export(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].ID != "rule-1" {
		t.Errorf("unexpected rules: %+v", rules)
	}
}

func TestDetectionsExportWithGroup(t *testing.T) {
	var gotQuery string
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		writeJSON(w, 200, []map[string]any{})
	}))
	defer cleanup()

	_, _ = client.Detections.Export(context.Background(), "production")
	if gotQuery != "format=json&group=production" {
		t.Errorf("query = %q, want format=json&group=production", gotQuery)
	}
}

func TestDetectionsExport_Unauthorized(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer cleanup()

	_, err := client.Detections.Export(context.Background(), "")
	if !errors.Is(err, craftedsignal.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestDetectionsImport(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/detections/import" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"success": true, "created": 1, "updated": 0, "unchanged": 0,
			"results": []map[string]any{
				{"id": "rule-1", "title": "Test", "action": "created", "version": 1},
			},
		})
	}))
	defer cleanup()

	atomic := true
	resp, err := client.Detections.Import(context.Background(), craftedsignal.ImportRequest{
		Rules:   []craftedsignal.Detection{{Title: "Test", Platform: "splunk"}},
		Message: "initial import",
		Mode:    "upsert",
		Atomic:  &atomic,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Created != 1 {
		t.Errorf("Created = %d, want 1", resp.Created)
	}
}

func TestDetectionsDeploy(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detections/deploy" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"deployed": 1, "failed": 0,
			"results": []map[string]any{{"id": "rule-1", "title": "T", "action": "deployed"}},
		})
	}))
	defer cleanup()

	resp, err := client.Detections.Deploy(context.Background(), []string{"rule-1"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Deployed != 1 {
		t.Errorf("Deployed = %d, want 1", resp.Deployed)
	}
}

func TestDetectionsGetSyncStatus(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detections/sync-status" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"rules": []map[string]any{{"id": "rule-1", "title": "T", "hash": "abc123", "version": 2}},
		})
	}))
	defer cleanup()

	status, err := client.Detections.GetSyncStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Rules) != 1 || status.Rules[0].ID != "rule-1" {
		t.Errorf("unexpected status: %+v", status)
	}
}

func TestDetectionsTest_HighLevel(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			// StartTests
			writeJSON(w, 200, map[string]any{
				"started": 1, "skipped": 0, "errors": 0,
				"results": []map[string]any{{"id": "rule-1", "title": "T", "action": "started"}},
			})
			return
		}
		// PollTests — pending first call, then complete
		pending := 1
		if calls > 2 {
			pending = 0
		}
		writeJSON(w, 200, map[string]any{
			"passed": 1 - pending, "failed": 0, "pending": pending,
			"results": []map[string]any{{"id": "rule-1", "title": "T", "test_status": "passing"}},
		})
	}))
	defer cleanup()

	status, err := client.Detections.Test(context.Background(), []string{"rule-1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status.Passed != 1 {
		t.Errorf("Passed = %d, want 1", status.Passed)
	}
}

func TestDetectionsGenerate_HighLevel(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"workflow_id": "wf-1", "status": "running"})
			return
		}
		// Poll — running first, then completed
		status := "running"
		rules := []map[string]any{}
		if calls > 2 {
			status = "completed"
			rules = []map[string]any{{"id": "rule-new", "title": "Generated", "platform": "splunk"}}
		}
		writeJSON(w, 200, map[string]any{"status": status, "rules": rules})
	}))
	defer cleanup()

	result, err := client.Detections.Generate(context.Background(),
		craftedsignal.GenerateRequest{Description: "detect PsExec", Platform: "splunk"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rules) != 1 {
		t.Errorf("Rules len = %d, want 1", len(result.Rules))
	}
}
