package craftedsignal

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
)

// BackoffFunc calculates the wait duration before retry attempt n (0-indexed).
type BackoffFunc func(attempt int) time.Duration

// ExponentialBackoff waits 2^n seconds between retries (1s, 2s, 4s, …).
var ExponentialBackoff BackoffFunc = func(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

// NoRetry disables retries.
var NoRetry BackoffFunc = func(_ int) time.Duration { return 0 }

type transport struct {
	token        Token
	baseURL      string
	httpClient   *http.Client
	maxRetries   int
	backoff      BackoffFunc
	logger       *slog.Logger
	verbose      bool
	pollInterval time.Duration
	userAgent    string
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *apiErrBody     `json:"error,omitempty"`
}

type apiErrBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// do executes an HTTP request with retry logic.
// The caller is responsible for passing resp to decode(), which closes the body.
func (t *transport) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var rawBody []byte
	if body != nil {
		var err error
		rawBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("craftedsignal: marshal request: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		if attempt > 0 {
			wait := t.backoff(attempt - 1)
			t.logDebug("retrying request",
				slog.String("method", method),
				slog.String("path", path),
				slog.Int("attempt", attempt),
				slog.Duration("wait", wait),
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		var reqBody io.Reader
		if rawBody != nil {
			reqBody = bytes.NewReader(rawBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, t.baseURL+path, reqBody)
		if err != nil {
			return nil, fmt.Errorf("craftedsignal: build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+t.token.value())
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if t.userAgent != "" {
			req.Header.Set("User-Agent", t.userAgent)
		}

		t.logDebug("request", slog.String("method", method), slog.String("path", path))

		resp, err := t.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("craftedsignal: %w", err)
			continue
		}

		t.logDebug("response",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", resp.StatusCode),
		)

		// Retry on 429 with Retry-After support
		if resp.StatusCode == http.StatusTooManyRequests && attempt < t.maxRetries {
			wait := parseRetryAfter(resp.Header.Get("Retry-After"), t.backoff(attempt))
			_ = resp.Body.Close()
			t.logger.Warn("rate limited",
				slog.Int("attempt", attempt+1),
				slog.Duration("retry_after", wait),
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		// Retry on 5xx (except last attempt)
		if resp.StatusCode >= 500 && attempt < t.maxRetries {
			_ = resp.Body.Close()
			lastErr = &Error{StatusCode: resp.StatusCode}
			continue
		}

		return resp, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("craftedsignal: request failed after %d attempts", t.maxRetries+1)
}

// decode reads the API envelope from resp and unmarshals Data into out.
// It always closes resp.Body. out may be nil to discard the data field.
func (t *transport) decode(resp *http.Response, out any) error {
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return fmt.Errorf("craftedsignal: read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	}

	if resp.StatusCode >= 400 {
		var env apiEnvelope
		if json.Unmarshal(body, &env) == nil && env.Error != nil {
			return &Error{Code: env.Error.Code, Message: env.Error.Message, StatusCode: resp.StatusCode}
		}
		return &Error{Code: "unexpected_error", Message: string(body), StatusCode: resp.StatusCode}
	}

	if out == nil {
		return nil
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("craftedsignal: decode envelope: %w", err)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("craftedsignal: decode data: %w", err)
	}
	return nil
}

func (t *transport) logDebug(msg string, args ...any) {
	if !t.verbose {
		return
	}
	t.logger.Debug(msg, args...)
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newInsecureTransport(base *http.Transport) *http.Transport {
	clone := base.Clone()
	clone.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	return clone
}

func parseRetryAfter(header string, fallback time.Duration) time.Duration {
	if header == "" {
		return fallback
	}
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
	if ts, err := http.ParseTime(header); err == nil {
		if d := time.Until(ts); d > 0 {
			return d
		}
	}
	return fallback
}
