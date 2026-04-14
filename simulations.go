package craftedsignal

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// SimulationsService manages simulation runs and coverage analysis.
type SimulationsService interface {
	// CreateRun creates a new simulation run.
	CreateRun(ctx context.Context, req CreateSimulationRequest) (*SimulationRun, error)
	// ListRuns returns all simulation runs.
	ListRuns(ctx context.Context) ([]SimulationRun, error)
	// GetRun returns a single simulation run by ID.
	GetRun(ctx context.Context, id string) (*SimulationRun, error)
	// DeleteRun deletes a simulation run.
	DeleteRun(ctx context.Context, id string) error
	// Coverage returns detection coverage across MITRE techniques.
	Coverage(ctx context.Context) (*CoverageReport, error)
	// Gaps returns MITRE techniques with no covering detection.
	Gaps(ctx context.Context) ([]CoverageGap, error)

	// Low-level async verification methods.
	StartVerify(ctx context.Context, id string) (*VerifyJob, error)
	PollVerify(ctx context.Context, id string) (*VerifyResult, error)
	// Verify triggers MITRE correlation for a run and polls until complete.
	Verify(ctx context.Context, id string, progress ProgressFunc) (*VerifyResult, error)
}

type simulationsService struct{ t *transport }

func (s *simulationsService) CreateRun(ctx context.Context, req CreateSimulationRequest) (*SimulationRun, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/simulations/runs", req)
	if err != nil {
		return nil, err
	}
	var result SimulationRun
	return &result, s.t.decode(resp, &result)
}

func (s *simulationsService) ListRuns(ctx context.Context) ([]SimulationRun, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/simulations/runs", nil)
	if err != nil {
		return nil, err
	}
	var result []SimulationRun
	return result, s.t.decode(resp, &result)
}

func (s *simulationsService) GetRun(ctx context.Context, id string) (*SimulationRun, error) {
	path := "/api/v1/simulations/runs/" + url.PathEscape(id)
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result SimulationRun
	return &result, s.t.decode(resp, &result)
}

func (s *simulationsService) DeleteRun(ctx context.Context, id string) error {
	path := "/api/v1/simulations/runs/" + url.PathEscape(id)
	resp, err := s.t.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}

func (s *simulationsService) Coverage(ctx context.Context) (*CoverageReport, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/simulations/coverage", nil)
	if err != nil {
		return nil, err
	}
	var result CoverageReport
	return &result, s.t.decode(resp, &result)
}

func (s *simulationsService) Gaps(ctx context.Context) ([]CoverageGap, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/simulations/gaps", nil)
	if err != nil {
		return nil, err
	}
	var result []CoverageGap
	return result, s.t.decode(resp, &result)
}

func (s *simulationsService) StartVerify(ctx context.Context, id string) (*VerifyJob, error) {
	path := "/api/v1/simulations/verify/" + url.PathEscape(id)
	resp, err := s.t.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	var result VerifyJob
	return &result, s.t.decode(resp, &result)
}

func (s *simulationsService) PollVerify(ctx context.Context, id string) (*VerifyResult, error) {
	path := "/api/v1/simulations/verify/" + url.PathEscape(id)
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result VerifyResult
	return &result, s.t.decode(resp, &result)
}

// Verify triggers MITRE correlation for a simulation run and polls until
// it reaches a terminal state or ctx is cancelled.
func (s *simulationsService) Verify(ctx context.Context, id string, progress ProgressFunc) (*VerifyResult, error) {
	if _, err := s.StartVerify(ctx, id); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(s.t.pollInterval):
		}
		result, err := s.PollVerify(ctx, id)
		if err != nil {
			return nil, err
		}
		if progress != nil {
			progress(result.Status, -1)
		}
		switch result.Status {
		case "completed":
			return result, nil
		case "failed", "error":
			return nil, &Error{Code: "verification_failed", Message: result.Error, StatusCode: 422}
		}
	}
}
