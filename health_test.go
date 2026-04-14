package craftedsignal_test

import (
	"context"
	"net/http"
	"testing"
)

func TestHealthCompanyMetrics(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health/company" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"total_rules": 50, "passing_rules": 45, "failing_rules": 5, "health_score": 0.9,
		})
	}))
	defer cleanup()

	m, err := client.Health.CompanyMetrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m.TotalRules != 50 {
		t.Errorf("TotalRules = %d, want 50", m.TotalRules)
	}
}

func TestHealthNoiseBudget(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health/noise-budget" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{"daily_budget": 100, "current_alerts": 30, "utilisation": 0.3})
	}))
	defer cleanup()

	nb, err := client.Health.NoiseBudget(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if nb.DailyBudget != 100 {
		t.Errorf("DailyBudget = %d, want 100", nb.DailyBudget)
	}
}

func TestHealthDeadRules(t *testing.T) {
	client, cleanup := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health/dead-rules" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, 200, []map[string]any{{"id": "rule-dead", "title": "Unused Rule"}})
	}))
	defer cleanup()

	rules, err := client.Health.DeadRules(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].ID != "rule-dead" {
		t.Errorf("unexpected rules: %+v", rules)
	}
}
