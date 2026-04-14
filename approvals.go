package craftedsignal

import (
	"context"
	"net/http"
	"net/url"
)

// ApprovalsService manages deployment approvals.
type ApprovalsService interface {
	// List returns all pending approvals for the workspace.
	List(ctx context.Context) ([]Approval, error)
	// Get returns a single approval by ID.
	Get(ctx context.Context, id string) (*Approval, error)
	// Approve approves a pending deployment.
	Approve(ctx context.Context, id string) error
	// Reject rejects a pending deployment.
	Reject(ctx context.Context, id string) error
}

type approvalsService struct{ t *transport }

func (s *approvalsService) List(ctx context.Context) ([]Approval, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/approvals", nil)
	if err != nil {
		return nil, err
	}
	var result []Approval
	return result, s.t.decode(resp, &result)
}

func (s *approvalsService) Get(ctx context.Context, id string) (*Approval, error) {
	path := "/api/v1/approvals/" + url.PathEscape(id)
	resp, err := s.t.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var result Approval
	return &result, s.t.decode(resp, &result)
}

func (s *approvalsService) Approve(ctx context.Context, id string) error {
	path := "/api/v1/approvals/" + url.PathEscape(id) + "/approve"
	resp, err := s.t.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	return s.t.decode(resp, nil)
}

func (s *approvalsService) Reject(ctx context.Context, id string) error {
	path := "/api/v1/approvals/" + url.PathEscape(id) + "/reject"
	resp, err := s.t.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	return s.t.decode(resp, nil)
}
