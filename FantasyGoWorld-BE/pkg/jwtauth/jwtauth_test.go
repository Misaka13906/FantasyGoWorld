package jwtauth

import (
	"testing"
)

func TestJWT(t *testing.T) {
	secret := "test-secret"
	uid := uint(123)
	expire := int64(3600)

	// Test GenerateToken
	token, err := GenerateToken(uid, secret, expire)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	// Test ParseToken
	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != uid {
		t.Errorf("expected uid %d, got %d", uid, claims.UserID)
	}

	// Test Invalid Secret
	_, err = ParseToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error with wrong secret, got nil")
	}

	// Test Expired Token
	expiredToken, err := GenerateToken(uid, secret, -1)
	if err != nil {
		t.Fatalf("GenerateToken failed for expired token: %v", err)
	}
	_, err = ParseToken(expiredToken, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}
