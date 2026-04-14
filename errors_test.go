package craftedsignal_test

import (
	"errors"
	"fmt"
	"testing"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

func TestErrorSentinels(t *testing.T) {
	err := &craftedsignal.Error{Code: "rule_not_found", Message: "not found", StatusCode: 404}

	if !errors.Is(err, craftedsignal.ErrNotFound) {
		t.Error("404 error should match ErrNotFound")
	}
	if errors.Is(err, craftedsignal.ErrUnauthorized) {
		t.Error("404 error should not match ErrUnauthorized")
	}
}

func TestErrorMessage(t *testing.T) {
	err := &craftedsignal.Error{Code: "bad_request", Message: "invalid query", StatusCode: 400}
	got := err.Error()
	if got != "[400 bad_request] invalid query" {
		t.Errorf("Error() = %q", got)
	}
}

func TestErrorsAs(t *testing.T) {
	var apiErr *craftedsignal.Error
	wrapped := fmt.Errorf("wrapped: %w", &craftedsignal.Error{StatusCode: 403, Code: "forbidden"})
	if !errors.As(wrapped, &apiErr) {
		t.Error("errors.As should unwrap *Error")
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
}
