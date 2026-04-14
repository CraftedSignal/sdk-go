// Package craftedsignal provides a Go client for the CraftedSignal API.
package craftedsignal

// Token is the bearer secret used to authenticate API requests.
// It redacts itself from string representations to prevent accidental
// leakage into logs, debug output, or serialised data.
type Token string

// String implements fmt.Stringer. Always returns "[REDACTED]".
func (t Token) String() string { return "[REDACTED]" }

// GoString implements fmt.GoStringer. Always returns a redacted representation.
func (t Token) GoString() string { return `craftedsignal.Token("[REDACTED]")` }

// MarshalJSON implements json.Marshaler. Always returns `"[REDACTED]"`.
func (t Token) MarshalJSON() ([]byte, error) { return []byte(`"[REDACTED]"`), nil }

// MarshalText implements encoding.TextMarshaler. Always returns `[REDACTED]`.
func (t Token) MarshalText() ([]byte, error) { return []byte("[REDACTED]"), nil }

// value returns the raw token for use in HTTP Authorization headers.
// This method is unexported and only callable within this package.
func (t Token) value() string { return string(t) }
