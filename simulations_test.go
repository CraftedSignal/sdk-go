package craftedsignal_test

import (
	"context"
	"net/http"
	"testing"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

func TestSimulationsCreateRun(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/simulations/runs" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"id": "run-1", "status": "planned", "technique_id": "T1078"})
	}))
	defer cleanup()

	run, err := client.Simulations.CreateRun(context.Background(), craftedsignal.CreateSimulationRequest{
		TechniqueID: "T1078",
		Adapter:     "atomic",
		Target:      "linux-host",
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.ID != "run-1" {
		t.Errorf("ID = %q, want run-1", run.ID)
	}
}

func TestSimulationsListRuns(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/simulations/runs" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, []map[string]any{{"id": "run-1"}, {"id": "run-2"}})
	}))
	defer cleanup()

	runs, err := client.Simulations.ListRuns(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Errorf("len = %d, want 2", len(runs))
	}
}

func TestSimulationsCoverage(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/simulations/coverage" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"total": 100, "covered": 42, "coverage": 0.42})
	}))
	defer cleanup()

	cov, err := client.Simulations.Coverage(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cov.Covered != 42 {
		t.Errorf("Covered = %d, want 42", cov.Covered)
	}
}

func TestSimulationsGaps(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, []map[string]any{
			{"technique_id": "T1059", "technique_name": "Command and Scripting Interpreter", "tactic": "execution"},
		})
	}))
	defer cleanup()

	gaps, err := client.Simulations.Gaps(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) != 1 || gaps[0].TechniqueID != "T1059" {
		t.Errorf("unexpected gaps: %+v", gaps)
	}
}

func TestSimulationsVerify_HighLevel(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"run_id": "run-1", "status": "correlating"})
			return
		}
		status := "correlating"
		if calls > 2 {
			status = "completed"
		}
		writeJSON(w, 200, map[string]any{"status": status, "results": []any{}})
	}))
	defer cleanup()

	result, err := client.Simulations.Verify(context.Background(), "run-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %q, want completed", result.Status)
	}
}

func TestSimulationsDeleteRun(t *testing.T) {
	var gotMethod, gotPath string
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer cleanup()

	err := client.Simulations.DeleteRun(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete || gotPath != "/api/v1/simulations/runs/run-1" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}
}

func TestSimulationsGetRun(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/simulations/runs/run-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"id": "run-1", "status": "completed"})
	}))
	defer cleanup()

	run, err := client.Simulations.GetRun(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if run.ID != "run-1" {
		t.Errorf("ID = %q", run.ID)
	}
}

func TestSimulationsStartVerify(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/simulations/verify/run-1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"run_id": "run-1", "status": "correlating"})
	}))
	defer cleanup()

	job, err := client.Simulations.StartVerify(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if job.RunID != "run-1" {
		t.Errorf("RunID = %q, want run-1", job.RunID)
	}
}

func TestSimulationsPollVerify(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/simulations/verify/run-1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"status": "completed", "results": []any{}})
	}))
	defer cleanup()

	result, err := client.Simulations.PollVerify(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %q, want completed", result.Status)
	}
}

func TestSimulationsVerify_Failure(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"run_id": "run-1", "status": "correlating"})
			return
		}
		writeJSON(w, 200, map[string]any{"status": "failed", "error": "no matching events"})
	}))
	defer cleanup()

	_, err := client.Simulations.Verify(context.Background(), "run-1", nil)
	if err == nil {
		t.Fatal("expected error on failed verification")
	}
}

func TestSimulationsVerify_WithProgress(t *testing.T) {
	calls := 0
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"run_id": "run-1", "status": "correlating"})
			return
		}
		status := "correlating"
		if calls > 2 {
			status = "completed"
		}
		writeJSON(w, 200, map[string]any{"status": status, "results": []any{}})
	}))
	defer cleanup()

	var progressCalls int
	progress := func(msg string, pct int) { progressCalls++ }
	result, err := client.Simulations.Verify(context.Background(), "run-1", progress)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" {
		t.Errorf("Status = %q, want completed", result.Status)
	}
	if progressCalls == 0 {
		t.Error("expected progress to be called at least once")
	}
}

func TestSimulationsVerify_ContextCancel(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			writeJSON(w, 200, map[string]any{"run_id": "run-1", "status": "correlating"})
			return
		}
		writeJSON(w, 200, map[string]any{"status": "correlating", "results": []any{}})
	}))
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.Simulations.Verify(ctx, "run-1", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
