package craftedsignal

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Client is the CraftedSignal API client.
// Create one with NewClient and use the service fields to call the API.
type Client struct {
	t           *transport
	Detections  DetectionsService
	Approvals   ApprovalsService
	Simulations SimulationsService
	Health      HealthService
	APIKeys     APIKeysService
}

// Option configures a Client.
type Option func(*transport)

// WithBaseURL overrides the API base URL. Default: "https://app.craftedsignal.io".
func WithBaseURL(url string) Option {
	return func(t *transport) { t.baseURL = url }
}

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(t *transport) { t.httpClient = c }
}

// WithRetry configures retry behaviour. Default: 3 retries with ExponentialBackoff.
func WithRetry(maxRetries int, backoff BackoffFunc) Option {
	return func(t *transport) {
		t.maxRetries = maxRetries
		t.backoff = backoff
	}
}

// WithInsecure disables TLS certificate verification. For development only.
func WithInsecure() Option {
	return func(t *transport) {
		base, ok := t.httpClient.Transport.(*http.Transport)
		if !ok {
			base = http.DefaultTransport.(*http.Transport)
		}
		t.httpClient.Transport = newInsecureTransport(base)
	}
}

// WithUserAgent sets a custom User-Agent header on all requests.
func WithUserAgent(ua string) Option {
	return func(t *transport) { t.userAgent = ua }
}

// WithLogger sets the slog.Logger. Default: slog.Default().
func WithLogger(l *slog.Logger) Option {
	return func(t *transport) { t.logger = l }
}

// WithVerbose enables DEBUG-level logging for requests, retries, and poll ticks.
func WithVerbose() Option {
	return func(t *transport) { t.verbose = true }
}

// WithPollInterval sets the cadence for async polling helpers. Default: 2s.
func WithPollInterval(d time.Duration) Option {
	return func(t *transport) { t.pollInterval = d }
}

// NewClient creates a new CraftedSignal API client authenticated with token.
// Returns an error if token is empty.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, errors.New("craftedsignal: token must not be empty")
	}

	tr := &transport{
		token:   Token(token),
		baseURL: "https://app.craftedsignal.io",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries:   3,
		backoff:      ExponentialBackoff,
		logger:       slog.Default(),
		pollInterval: 2 * time.Second,
	}

	for _, opt := range opts {
		opt(tr)
	}

	c := &Client{t: tr}
	c.Detections = &detectionsService{t: tr}
	c.Approvals = &approvalsService{t: tr}
	c.Simulations = &simulationsService{t: tr}
	c.Health = &healthService{t: tr}
	c.APIKeys = &apiKeysService{t: tr}
	return c, nil
}

// Me returns authentication information for the current token.
func (c *Client) Me(ctx context.Context) (*Me, error) {
	resp, err := c.t.do(ctx, http.MethodGet, "/api/v1/me", nil)
	if err != nil {
		return nil, err
	}
	var result Me
	if err := c.t.decode(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
