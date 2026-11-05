package craftedsignal

import (
	"context"
	"net/http"
)

// LibraryService manages local reusable library templates and guides.
type LibraryService interface {
	Export(ctx context.Context) (*LibraryExport, error)
	GetSyncStatus(ctx context.Context) (*LibrarySyncStatus, error)
	Import(ctx context.Context, req LibraryImportRequest) (*LibraryImportResponse, error)
}

type libraryService struct{ t *transport }

func (s *libraryService) Export(ctx context.Context) (*LibraryExport, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/library/export?format=json", nil)
	if err != nil {
		return nil, err
	}
	var result LibraryExport
	return &result, s.t.decode(resp, &result)
}

func (s *libraryService) GetSyncStatus(ctx context.Context) (*LibrarySyncStatus, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/library/sync-status", nil)
	if err != nil {
		return nil, err
	}
	var result LibrarySyncStatus
	return &result, s.t.decode(resp, &result)
}

func (s *libraryService) Import(ctx context.Context, req LibraryImportRequest) (*LibraryImportResponse, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/library/import", req)
	if err != nil {
		return nil, err
	}
	var result LibraryImportResponse
	result.StatusCode = resp.StatusCode
	return &result, s.t.decode(resp, &result)
}
