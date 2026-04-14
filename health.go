package craftedsignal

import (
	"context"
	"net/http"
)

// HealthService provides company-wide and per-rule health metrics.
type HealthService interface {
	// CompanyMetrics returns overall detection health for the workspace.
	CompanyMetrics(ctx context.Context) (*HealthMetrics, error)
	// NoiseBudget returns alert fatigue budget metrics.
	NoiseBudget(ctx context.Context) (*NoiseBudget, error)
	// DeadRules returns detection rules with no recent activity.
	DeadRules(ctx context.Context) ([]Detection, error)
}

type healthService struct{ t *transport }

func (s *healthService) CompanyMetrics(ctx context.Context) (*HealthMetrics, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/health/company", nil)
	if err != nil {
		return nil, err
	}
	var result HealthMetrics
	return &result, s.t.decode(resp, &result)
}

func (s *healthService) NoiseBudget(ctx context.Context) (*NoiseBudget, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/health/noise-budget", nil)
	if err != nil {
		return nil, err
	}
	var result NoiseBudget
	return &result, s.t.decode(resp, &result)
}

func (s *healthService) DeadRules(ctx context.Context) ([]Detection, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/health/dead-rules", nil)
	if err != nil {
		return nil, err
	}
	var result []Detection
	return result, s.t.decode(resp, &result)
}
