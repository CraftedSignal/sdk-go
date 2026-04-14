package craftedsignal

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DetectionsService manages detection rules.
type DetectionsService interface {
	Export(ctx context.Context, group string) ([]Detection, error)
	GetSyncStatus(ctx context.Context) (*SyncStatus, error)
	ExportOne(ctx context.Context, id string) (*Detection, error)
	Import(ctx context.Context, req ImportRequest) (*ImportResponse, error)
	Diff(ctx context.Context, id string, local Detection) (*DiffResult, error)
	Deploy(ctx context.Context, ids []string, overrideTests bool) (*DeployResponse, error)
	Health(ctx context.Context, id string) (*DetectionHealth, error)

	// Low-level async test methods.
	StartTests(ctx context.Context, ids []string) (*TestJob, error)
	PollTests(ctx context.Context, ids []string) (*TestResponse, error)
	// Test triggers tests and polls until all complete or ctx is cancelled.
	// progress is called on each poll tick; nil is safe.
	Test(ctx context.Context, ids []string, progress ProgressFunc) (*TestResponse, error)

	// Low-level async AI generation methods.
	StartGenerate(ctx context.Context, req GenerateRequest) (*GenerateJob, error)
	PollGenerate(ctx context.Context, workflowID string) (*GenerateResult, error)
	// Generate starts AI rule generation and polls until complete.
	Generate(ctx context.Context, req GenerateRequest, progress ProgressFunc) (*GenerateResult, error)
}

// DetectionHealth holds the health score for a single detection rule.
type DetectionHealth struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

type detectionsService struct{ t *transport }

func (s *detectionsService) Export(ctx context.Context, group string) ([]Detection, error) {
	path := "/api/v1/detections/export?format=json"
	if group != "" {
		path += "&group=" + url.QueryEscape(group)
	}
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result []Detection
	return result, s.t.decode(resp, &result)
}

func (s *detectionsService) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/detections/sync-status", nil)
	if err != nil {
		return nil, err
	}
	var result SyncStatus
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) ExportOne(ctx context.Context, id string) (*Detection, error) {
	path := "/api/v1/detections/" + url.PathEscape(id) + "/export"
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result Detection
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) Import(ctx context.Context, req ImportRequest) (*ImportResponse, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/detections/import", req)
	if err != nil {
		return nil, err
	}
	var result ImportResponse
	result.StatusCode = resp.StatusCode
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) Diff(ctx context.Context, id string, local Detection) (*DiffResult, error) {
	path := "/api/v1/detections/" + url.PathEscape(id) + "/diff"
	resp, err := s.t.do(ctx, http.MethodPost, path, local)
	if err != nil {
		return nil, err
	}
	var result DiffResult
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) Deploy(ctx context.Context, ids []string, overrideTests bool) (*DeployResponse, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/detections/deploy", DeployRequest{
		DetectionIDs:  ids,
		OverrideTests: overrideTests,
	})
	if err != nil {
		return nil, err
	}
	var result DeployResponse
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) Health(ctx context.Context, id string) (*DetectionHealth, error) {
	path := "/api/v1/detections/" + url.PathEscape(id) + "/health"
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result DetectionHealth
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) StartTests(ctx context.Context, ids []string) (*TestJob, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/detections/test",
		map[string]any{"detection_ids": ids})
	if err != nil {
		return nil, err
	}
	var result TestJob
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) PollTests(ctx context.Context, ids []string) (*TestResponse, error) {
	q := url.Values{}
	for _, id := range ids {
		q.Add("ids", id)
	}
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/detections/test-status?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	var result TestResponse
	return &result, s.t.decode(resp, &result)
}

// Test starts tests and polls until all reach a terminal state or ctx is cancelled.
func (s *detectionsService) Test(ctx context.Context, ids []string, progress ProgressFunc) (*TestResponse, error) {
	job, err := s.StartTests(ctx, ids)
	if err != nil {
		return nil, err
	}

	var startedIDs []string
	for _, r := range job.Results {
		if r.Action == "started" {
			startedIDs = append(startedIDs, r.ID)
		}
	}
	if len(startedIDs) == 0 {
		return &TestResponse{Results: []TestResult{}}, nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(s.t.pollInterval):
		}

		status, err := s.PollTests(ctx, startedIDs)
		if err != nil {
			return nil, err
		}
		if progress != nil {
			total := len(startedIDs)
			done := total - status.Pending
			pct := -1
			if total > 0 {
				pct = (done * 100) / total
			}
			progress(fmt.Sprintf("%d/%d complete", done, total), pct)
		}
		if status.Pending == 0 {
			return status, nil
		}
	}
}

func (s *detectionsService) StartGenerate(ctx context.Context, req GenerateRequest) (*GenerateJob, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/detections/generate", req)
	if err != nil {
		return nil, err
	}
	var result GenerateJob
	return &result, s.t.decode(resp, &result)
}

func (s *detectionsService) PollGenerate(ctx context.Context, workflowID string) (*GenerateResult, error) {
	path := "/api/v1/detections/generate/status/" + url.PathEscape(workflowID)
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result GenerateResult
	return &result, s.t.decode(resp, &result)
}

// Generate starts AI rule generation and polls until the workflow reaches a terminal state.
func (s *detectionsService) Generate(ctx context.Context, req GenerateRequest, progress ProgressFunc) (*GenerateResult, error) {
	job, err := s.StartGenerate(ctx, req)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(s.t.pollInterval):
		}

		result, err := s.PollGenerate(ctx, job.WorkflowID)
		if err != nil {
			return nil, err
		}
		if progress != nil {
			progress(result.Status, -1)
		}
		switch result.Status {
		case "completed":
			return result, nil
		case "failed", "cancelled", "error":
			return nil, &Error{Code: "generation_failed", Message: result.Error, StatusCode: 422}
		}
	}
}
