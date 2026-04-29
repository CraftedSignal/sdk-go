package craftedsignal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DefaultFeedBaseURL is the public threat feed origin.
const DefaultFeedBaseURL = "https://feed.craftedsignal.io"

// Feed is a JSON Feed response from the public threat feed. One Feed is
// returned per filter (severity, type, product, vendor, actor, tag) and
// for the all-briefs view.
type Feed struct {
	Version     string     `json:"version"`
	Title       string     `json:"title"`
	SiteURL     string     `json:"site_url"`
	FeedURL     string     `json:"feed_url"`
	Description string     `json:"description"`
	Items       []FeedItem `json:"items"`
}

// FeedItem is one brief in a feed listing.
type FeedItem struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Date       time.Time `json:"date"`
	Type       string    `json:"type"`
	Severities []string  `json:"severities"`
	Products   []string  `json:"products"`
	Vendors    []string  `json:"vendors"`
	Actors     []string  `json:"actors"`
	Tags       []string  `json:"tags"`
	Exploited  bool      `json:"exploited"`
	CVEs       []FeedCVE `json:"cves"`
}

// FeedCVE is a CVE reference inside a feed item.
type FeedCVE struct {
	ID   string  `json:"id"`
	CVSS float64 `json:"cvss"`
}

// FeedClient reads the public threat feed at feed.craftedsignal.io. It
// is unauthenticated and rate-limited only by the static-site CDN, so a
// nil-token program can construct one without going through NewClient.
type FeedClient struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// FeedOption configures a FeedClient.
type FeedOption func(*FeedClient)

// WithFeedBaseURL overrides the feed origin. Useful for staging
// mirrors or self-hosted forks. Default: DefaultFeedBaseURL.
func WithFeedBaseURL(u string) FeedOption {
	return func(c *FeedClient) { c.baseURL = u }
}

// WithFeedHTTPClient supplies a custom *http.Client (for timeouts,
// retries, or a transport that adds tracing).
func WithFeedHTTPClient(h *http.Client) FeedOption {
	return func(c *FeedClient) { c.httpClient = h }
}

// WithFeedUserAgent sets a custom User-Agent header on every request.
func WithFeedUserAgent(ua string) FeedOption {
	return func(c *FeedClient) { c.userAgent = ua }
}

// NewFeedClient returns a client for the public threat feed.
func NewFeedClient(opts ...FeedOption) *FeedClient {
	c := &FeedClient{
		baseURL:    DefaultFeedBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "craftedsignal-sdk-go",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Latest fetches the all-briefs feed.
func (c *FeedClient) Latest(ctx context.Context) (*Feed, error) {
	return c.fetch(ctx, "/feed.json")
}

// BySeverity fetches the feed filtered by severity.
// term must be one of: critical, high, medium, low, rumour.
func (c *FeedClient) BySeverity(ctx context.Context, term string) (*Feed, error) {
	return c.fetchTaxonomy(ctx, "severities", term)
}

// ByType fetches the feed filtered by brief type.
// term must be one of: threat, coverage, advisory, rumour.
func (c *FeedClient) ByType(ctx context.Context, term string) (*Feed, error) {
	return c.fetchTaxonomy(ctx, "types", term)
}

// ByProduct fetches the feed for a specific product slug.
// Slugs are lowercase, hyphenated forms of the product name as it
// appears in Items[i].Products.
func (c *FeedClient) ByProduct(ctx context.Context, slug string) (*Feed, error) {
	return c.fetchTaxonomy(ctx, "products", slug)
}

// ByVendor fetches the feed for a specific vendor slug.
func (c *FeedClient) ByVendor(ctx context.Context, slug string) (*Feed, error) {
	return c.fetchTaxonomy(ctx, "vendors", slug)
}

// ByActor fetches the feed for a specific threat-actor slug.
func (c *FeedClient) ByActor(ctx context.Context, slug string) (*Feed, error) {
	return c.fetchTaxonomy(ctx, "actors", slug)
}

// ByTag fetches the feed for a specific tag slug.
func (c *FeedClient) ByTag(ctx context.Context, slug string) (*Feed, error) {
	return c.fetchTaxonomy(ctx, "tags", slug)
}

func (c *FeedClient) fetchTaxonomy(ctx context.Context, plural, term string) (*Feed, error) {
	if term == "" {
		return nil, fmt.Errorf("feed: %s term must not be empty", plural)
	}
	return c.fetch(ctx, fmt.Sprintf("/%s/%s/feed.json", plural, url.PathEscape(term)))
}

func (c *FeedClient) fetch(ctx context.Context, path string) (*Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("feed: build request: %w", err)
	}
	req.Header.Set("Accept", "application/feed+json, application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("feed: %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("feed: %s: read body: %w", path, err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("feed: %s: HTTP %d: %s", path, resp.StatusCode, string(body))
	}

	var feed Feed
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("feed: %s: decode: %w", path, err)
	}
	return &feed, nil
}
