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

func TestDetectionsExportOne(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detections/rule-1/export" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"id": "rule-1", "title": "T", "platform": "splunk"})
	}))
	defer cleanup()

	d, err := client.Detections.ExportOne(context.Background(), "rule-1")
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != "rule-1" {
		t.Errorf("ID = %q", d.ID)
	}
}

func TestDetectionsDiff(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/detections/rule-1/diff" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"has_diff": true, "diff": "--- a\n+++ b"})
	}))
	defer cleanup()

	result, err := client.Detections.Diff(context.Background(), "rule-1",
		craftedsignal.Detection{Title: "Updated", Platform: "splunk"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.HasDiff {
		t.Error("expected HasDiff = true")
	}
}

func TestDetectionsHealth(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detections/rule-1/health" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"id": "rule-1", "score": 0.95})
	}))
	defer cleanup()

	h, err := client.Detections.Health(context.Background(), "rule-1")
	if err != nil {
		t.Fatal(err)
	}
	if h.Score != 0.95 {
		t.Errorf("Score = %v, want 0.95", h.Score)
	}
}

func TestDetectionsStartTests(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"started": 1, "skipped": 0, "errors": 0,
			"results": []map[string]any{{"id": "r1", "title": "T", "action": "started"}},
		})
	}))
	defer cleanup()

	job, err := client.Detections.StartTests(context.Background(), []string{"r1"})
	if err != nil {
		t.Fatal(err)
	}
	if job.Started != 1 {
		t.Errorf("Started = %d, want 1", job.Started)
	}
}

func TestDetectionsPollTests(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"passed": 1, "failed": 0, "pending": 0,
			"results": []map[string]any{{"id": "r1", "test_status": "passing"}},
		})
	}))
	defer cleanup()

	resp, err := client.Detections.PollTests(context.Background(), []string{"r1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Passed != 1 {
		t.Errorf("Passed = %d, want 1", resp.Passed)
	}
}

func TestDetectionsStartGenerate(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"workflow_id": "wf-1", "status": "running"})
	}))
	defer cleanup()

	job, err := client.Detections.StartGenerate(context.Background(),
		craftedsignal.GenerateRequest{Description: "test", Platform: "splunk"})
	if err != nil {
		t.Fatal(err)
	}
	if job.WorkflowID != "wf-1" {
		t.Errorf("WorkflowID = %q, want wf-1", job.WorkflowID)
	}
}

func TestDetectionsPollGenerate(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detections/generate/status/wf-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"status": "running", "progress": "analyzing"})
	}))
	defer cleanup()

	result, err := client.Detections.PollGenerate(context.Background(), "wf-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "running" {
		t.Errorf("Status = %q, want running", result.Status)
	}
}

func TestDetectionsGenerate_Failure(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"workflow_id": "wf-1", "status": "running"})
			return
		}
		writeJSON(w, 200, map[string]any{"status": "failed", "error": "context limit exceeded"})
	}))
	defer cleanup()

	_, err := client.Detections.Generate(context.Background(),
		craftedsignal.GenerateRequest{Description: "test", Platform: "splunk"}, nil)
	if err == nil {
		t.Fatal("expected error on failed generation")
	}
}

func TestDetectionsTest_NoStarted(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"started": 0, "skipped": 1, "errors": 0,
			"results": []map[string]any{{"id": "r1", "title": "T", "action": "skipped"}},
		})
	}))
	defer cleanup()

	resp, err := client.Detections.Test(context.Background(), []string{"r1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestDetectionsTest_WithProgress(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{
				"started": 1, "skipped": 0, "errors": 0,
				"results": []map[string]any{{"id": "rule-1", "title": "T", "action": "started"}},
			})
			return
		}
		pending := 0
		if calls <= 2 {
			pending = 1
		}
		writeJSON(w, 200, map[string]any{
			"passed": 1 - pending, "failed": 0, "pending": pending,
			"results": []map[string]any{{"id": "rule-1", "title": "T", "test_status": "passing"}},
		})
	}))
	defer cleanup()

	var progressCalls int
	progress := func(msg string, pct int) { progressCalls++ }
	status, err := client.Detections.Test(context.Background(), []string{"rule-1"}, progress)
	if err != nil {
		t.Fatal(err)
	}
	if status.Passed != 1 {
		t.Errorf("Passed = %d, want 1", status.Passed)
	}
	if progressCalls == 0 {
		t.Error("expected progress to be called at least once")
	}
}

func TestDetectionsGenerate_WithProgress(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"workflow_id": "wf-1", "status": "running"})
			return
		}
		status := "running"
		rules := []map[string]any{}
		if calls > 2 {
			status = "completed"
			rules = []map[string]any{{"id": "rule-new", "title": "Generated", "platform": "splunk"}}
		}
		writeJSON(w, 200, map[string]any{"status": status, "rules": rules})
	}))
	defer cleanup()

	var progressCalls int
	progress := func(msg string, pct int) { progressCalls++ }
	result, err := client.Detections.Generate(context.Background(),
		craftedsignal.GenerateRequest{Description: "detect PsExec", Platform: "splunk"},
		progress,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rules) != 1 {
		t.Errorf("Rules len = %d, want 1", len(result.Rules))
	}
	if progressCalls == 0 {
		t.Error("expected progress to be called at least once")
	}
}

func TestDetectionsTest_ContextCancel(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{
				"started": 1, "skipped": 0, "errors": 0,
				"results": []map[string]any{{"id": "rule-1", "title": "T", "action": "started"}},
			})
			return
		}
		// Always return pending so we can cancel
		writeJSON(w, 200, map[string]any{
			"passed": 0, "failed": 0, "pending": 1,
			"results": []map[string]any{{"id": "rule-1", "test_status": "pending"}},
		})
	}))
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately after the first poll

	_, err := client.Detections.Test(ctx, []string{"rule-1"}, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestDetectionsGenerate_ContextCancel(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"workflow_id": "wf-1", "status": "running"})
			return
		}
		writeJSON(w, 200, map[string]any{"status": "running"})
	}))
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.Detections.Generate(ctx,
		craftedsignal.GenerateRequest{Description: "test", Platform: "splunk"}, nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
