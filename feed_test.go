package craftedsignal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const sampleFeedJSON = `{
  "version": "1.0",
  "title": "CraftedSignal Threat Feed",
  "site_url": "https://feed.craftedsignal.io/",
  "feed_url": "https://feed.craftedsignal.io/feed.json",
  "description": "test",
  "items": [
    {
      "id": "/briefs/2026-04-some-cve/",
      "url": "https://feed.craftedsignal.io/briefs/2026-04-some-cve/",
      "title": "Some CVE",
      "summary": "Test summary",
      "date": "2026-04-28T12:34:56Z",
      "type": "threat",
      "severities": ["critical"],
      "products": ["FortiGate"],
      "vendors": ["Fortinet"],
      "actors": [],
      "tags": ["cve"],
      "exploited": true,
      "cves": [{"id": "CVE-2026-1234", "cvss": 9.8}]
    }
  ]
}`

func newFeedTestServer(t *testing.T, paths map[string]string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for p, body := range paths {
		body := body
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/feed+json")
			_, _ = w.Write([]byte(body))
		})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestFeedClient_Latest(t *testing.T) {
	srv := newFeedTestServer(t, map[string]string{
		"/feed.json": sampleFeedJSON,
	})

	c := NewFeedClient(WithFeedBaseURL(srv.URL))
	feed, err := c.Latest(context.Background())
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}

	if feed.Version != "1.0" {
		t.Errorf("Version: got %q, want 1.0", feed.Version)
	}
	if len(feed.Items) != 1 {
		t.Fatalf("Items: got %d, want 1", len(feed.Items))
	}
	item := feed.Items[0]
	if item.Title != "Some CVE" {
		t.Errorf("Title: got %q", item.Title)
	}
	if !item.Date.Equal(time.Date(2026, 4, 28, 12, 34, 56, 0, time.UTC)) {
		t.Errorf("Date: got %v", item.Date)
	}
	if !item.Exploited {
		t.Error("Exploited: got false, want true")
	}
	if len(item.CVEs) != 1 || item.CVEs[0].ID != "CVE-2026-1234" || item.CVEs[0].CVSS != 9.8 {
		t.Errorf("CVEs: got %+v", item.CVEs)
	}
	if len(item.Severities) != 1 || item.Severities[0] != "critical" {
		t.Errorf("Severities: got %v", item.Severities)
	}
}

func TestFeedClient_TaxonomyPaths(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		call     func(*FeedClient) (*Feed, error)
		wantTerm string
	}{
		{"BySeverity", "/severities/critical/feed.json", func(c *FeedClient) (*Feed, error) {
			return c.BySeverity(context.Background(), "critical")
		}, "critical"},
		{"ByType", "/types/threat/feed.json", func(c *FeedClient) (*Feed, error) {
			return c.ByType(context.Background(), "threat")
		}, "threat"},
		{"ByProduct", "/products/fortigate/feed.json", func(c *FeedClient) (*Feed, error) {
			return c.ByProduct(context.Background(), "fortigate")
		}, "fortigate"},
		{"ByVendor", "/vendors/fortinet/feed.json", func(c *FeedClient) (*Feed, error) {
			return c.ByVendor(context.Background(), "fortinet")
		}, "fortinet"},
		{"ByActor", "/actors/apt29/feed.json", func(c *FeedClient) (*Feed, error) {
			return c.ByActor(context.Background(), "apt29")
		}, "apt29"},
		{"ByTag", "/tags/cve/feed.json", func(c *FeedClient) (*Feed, error) {
			return c.ByTag(context.Background(), "cve")
		}, "cve"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newFeedTestServer(t, map[string]string{
				tt.path: sampleFeedJSON,
			})
			c := NewFeedClient(WithFeedBaseURL(srv.URL))
			feed, err := tt.call(c)
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			if len(feed.Items) != 1 {
				t.Errorf("%s: items=%d", tt.name, len(feed.Items))
			}
		})
	}
}

func TestFeedClient_EmptyTermRejected(t *testing.T) {
	c := NewFeedClient()
	if _, err := c.BySeverity(context.Background(), ""); err == nil {
		t.Error("BySeverity(\"\") should error")
	}
	if _, err := c.ByProduct(context.Background(), ""); err == nil {
		t.Error("ByProduct(\"\") should error")
	}
}

func TestFeedClient_HTTPErrorPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	c := NewFeedClient(WithFeedBaseURL(srv.URL))
	_, err := c.Latest(context.Background())
	if err == nil {
		t.Fatal("expected error from 503")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error should mention status: %v", err)
	}
}

func TestFeedClient_UserAgentSent(t *testing.T) {
	got := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/feed+json")
		_, _ = w.Write([]byte(sampleFeedJSON))
	}))
	t.Cleanup(srv.Close)

	c := NewFeedClient(WithFeedBaseURL(srv.URL), WithFeedUserAgent("my-app/1.2"))
	if _, err := c.Latest(context.Background()); err != nil {
		t.Fatalf("Latest: %v", err)
	}
	select {
	case ua := <-got:
		if ua != "my-app/1.2" {
			t.Errorf("User-Agent: got %q", ua)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no request received")
	}
}
