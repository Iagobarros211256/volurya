package middleware

import (
	"encoding/hex"
	"testing"
)

func TestGenerateCSRFToken_Length(t *testing.T) {
	token, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("expected 64 chars, got %d", len(token))
	}
}

func TestGenerateCSRFToken_IsHex(t *testing.T) {
	token, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := hex.DecodeString(token); err != nil {
		t.Fatalf("token is not valid hex: %v", err)
	}
}

func TestGenerateCSRFToken_IsRandom(t *testing.T) {
	token1, _ := generateCSRFToken()
	token2, _ := generateCSRFToken()
	if token1 == token2 {
		t.Fatal("two generated tokens should not be equal")
	}
}
