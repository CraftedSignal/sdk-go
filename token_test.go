package craftedsignal_test

import (
	"encoding/json"
	"fmt"
	"testing"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

func TestTokenRedaction(t *testing.T) {
	tok := craftedsignal.Token("super-secret-key")

	if got := fmt.Sprintf("%s", tok); got != "[REDACTED]" {
		t.Errorf("String() = %q, want [REDACTED]", got)
	}
	if got := fmt.Sprintf("%v", tok); got != "[REDACTED]" {
		t.Errorf("%%v = %q, want [REDACTED]", got)
	}
	if got := fmt.Sprintf("%#v", tok); got == "super-secret-key" {
		t.Errorf("GoString() leaked token")
	}

	b, err := json.Marshal(tok)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"[REDACTED]"` {
		t.Errorf("MarshalJSON = %s, want \"[REDACTED]\"", b)
	}
}

func TestTokenMarshalText(t *testing.T) {
	tok := craftedsignal.Token("super-secret-key")
	b, err := tok.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "[REDACTED]" {
		t.Errorf("MarshalText = %q, want [REDACTED]", string(b))
	}
}
