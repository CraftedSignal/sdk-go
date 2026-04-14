package craftedsignal

import "fmt"

// Error represents an API error response from the CraftedSignal platform.
type Error struct {
	// Code is the machine-readable API error code, e.g. "rule_not_found".
	Code string
	// Message is the human-readable error description.
	Message string
	// StatusCode is the HTTP status code.
	StatusCode int
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d %s] %s", e.StatusCode, e.Code, e.Message)
}

// Is enables errors.Is() matching against sentinel error variables.
// Matching is based on StatusCode (if non-zero) and Code (if non-empty).
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	if t.StatusCode != 0 && t.StatusCode != e.StatusCode {
		return false
	}
	if t.Code != "" && t.Code != e.Code {
		return false
	}
	return true
}

// Sentinel errors for common HTTP status codes.
// Use with errors.Is: errors.Is(err, craftedsignal.ErrNotFound)
var (
	ErrUnauthorized = &Error{StatusCode: 401}
	ErrForbidden    = &Error{StatusCode: 403}
	ErrNotFound     = &Error{StatusCode: 404}
	ErrRateLimited  = &Error{StatusCode: 429}
	ErrServerError  = &Error{StatusCode: 500}
)
