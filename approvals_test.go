package craftedsignal_test

import (
	"context"
	"net/http"
	"testing"
)

func TestApprovalsList(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/approvals" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, []map[string]any{
			{"id": "appr-1", "status": "pending", "rule_id": "rule-1"},
		})
	}))
	defer cleanup()

	approvals, err := client.Approvals.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(approvals) != 1 || approvals[0].ID != "appr-1" {
		t.Errorf("unexpected approvals: %+v", approvals)
	}
}

func TestApprovalsGet(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/approvals/appr-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"id": "appr-1", "status": "pending"})
	}))
	defer cleanup()

	a, err := client.Approvals.Get(context.Background(), "appr-1")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != "appr-1" {
		t.Errorf("ID = %q, want appr-1", a.ID)
	}
}

func TestApprovalsApprove(t *testing.T) {
	var gotMethod, gotPath string
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		writeJSON(w, 200, map[string]any{})
	}))
	defer cleanup()

	err := client.Approvals.Approve(context.Background(), "appr-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost || gotPath != "/api/v1/approvals/appr-1/approve" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}
}

func TestApprovalsReject(t *testing.T) {
	var gotPath string
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		writeJSON(w, 200, map[string]any{})
	}))
	defer cleanup()

	err := client.Approvals.Reject(context.Background(), "appr-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/approvals/appr-1/reject" {
		t.Errorf("path = %s", gotPath)
	}
}
