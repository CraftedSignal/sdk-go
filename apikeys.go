package craftedsignal

import (
	"context"
	"net/http"
	"net/url"
)

// APIKeysService manages API keys. Requires admin scope.
type APIKeysService interface {
	// List returns all API keys for the workspace.
	List(ctx context.Context) ([]APIKey, error)
	// Create creates a new API key. The plaintext key is only returned here.
	Create(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyWithSecret, error)
	// Revoke permanently deletes an API key by ID.
	Revoke(ctx context.Context, id string) error
}

type apiKeysService struct{ t *transport }

func (s *apiKeysService) List(ctx context.Context) ([]APIKey, error) {
	resp, err := s.t.do(ctx, http.MethodGet, "/api/v1/api-keys", nil)
	if err != nil {
		return nil, err
	}
	var result []APIKey
	return result, s.t.decode(resp, &result)
}

func (s *apiKeysService) Create(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyWithSecret, error) {
	resp, err := s.t.do(ctx, http.MethodPost, "/api/v1/api-keys", req)
	if err != nil {
		return nil, err
	}
	var result APIKeyWithSecret
	return &result, s.t.decode(resp, &result)
}

func (s *apiKeysService) Revoke(ctx context.Context, id string) error {
	path := "/api/v1/api-keys/" + url.PathEscape(id)
	resp, err := s.t.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
