package auth

import (
	"testing"
	"time"
)

func TestPassword_HashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "correct-horse-battery" {
		t.Fatal("hash must not equal plaintext")
	}
	if !VerifyPassword(hash, "correct-horse-battery") {
		t.Error("VerifyPassword should accept the correct password")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Error("VerifyPassword should reject a wrong password")
	}
}

func newTM() *TokenManager {
	return NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
}

func TestAccessToken_RoundTrip(t *testing.T) {
	tm := newTM()
	tok, exp, err := tm.GenerateAccessToken("user-1", "acct-1", "user")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	if !exp.After(time.Now()) {
		t.Error("expiry should be in the future")
	}
	claims, err := tm.ParseAccessToken(tok)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if claims.UserID != "user-1" || claims.AccountID != "acct-1" || claims.Role != "user" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestAccessToken_RejectsWrongSecret(t *testing.T) {
	tok, _, _ := newTM().GenerateAccessToken("u", "a", "user")
	other := NewTokenManager("different-secret", 15*time.Minute, 24*time.Hour)
	if _, err := other.ParseAccessToken(tok); err == nil {
		t.Fatal("token signed with a different secret must be rejected")
	}
}

func TestAccessToken_RejectsExpired(t *testing.T) {
	tm := NewTokenManager("test-secret", -1*time.Minute, 24*time.Hour) // already expired
	tok, _, _ := tm.GenerateAccessToken("u", "a", "user")
	if _, err := tm.ParseAccessToken(tok); err == nil {
		t.Fatal("expired token must be rejected")
	}
}

func TestAccessToken_RejectsGarbage(t *testing.T) {
	if _, err := newTM().ParseAccessToken("not.a.jwt"); err == nil {
		t.Fatal("garbage token must be rejected")
	}
}

func TestRefreshToken_HashStableAndOpaque(t *testing.T) {
	tm := newTM()
	rt, err := tm.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if rt.Plaintext == "" || rt.Hash == "" {
		t.Fatal("refresh token and hash must be non-empty")
	}
	if rt.Hash == rt.Plaintext {
		t.Error("stored hash must differ from the plaintext token")
	}
	if HashToken(rt.Plaintext) != rt.Hash {
		t.Error("HashToken must reproduce the stored hash for lookups")
	}
	if !rt.ExpiresAt.After(time.Now()) {
		t.Error("refresh token expiry should be in the future")
	}

	// Two generations must differ (randomness).
	rt2, _ := tm.GenerateRefreshToken()
	if rt.Plaintext == rt2.Plaintext {
		t.Error("refresh tokens must be unique per generation")
	}
}
