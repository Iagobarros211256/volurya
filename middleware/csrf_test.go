package middleware

import (
	"testing"
)

func TestGenerateCSRFToken(t *testing.T) {
	sessionID := "session123"
	token1 := GenerateCSRFToken(sessionID)
	token2 := GenerateCSRFToken(sessionID)

	if token1 != token2 {
		t.Fatalf("same session should generate same token")
	}

	if len(token1) != 64 {
		t.Fatalf("token should be 64 chars (SHA256 hex), got %d", len(token1))
	}
}

func TestValidateCSRFToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{"valid token", GenerateCSRFToken("test"), true},
		{"invalid - empty", "", false},
		{"invalid - too short", "abc", false},
		{"invalid - not hex", "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", false},
		{"invalid - mixed case ok but must be hex", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isValidCSRFToken(test.token)
			if result != test.valid {
				t.Fatalf("expected %v, got %v", test.valid, result)
			}
		})
	}
}

func TestCSRFTokenDifferentSessions(t *testing.T) {
	token1 := GenerateCSRFToken("session1")
	token2 := GenerateCSRFToken("session2")

	if token1 == token2 {
		t.Fatalf("different sessions should generate different tokens")
	}
}
