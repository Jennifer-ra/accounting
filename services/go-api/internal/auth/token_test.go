package auth

import (
	"testing"
	"time"
)

func TestSignAndParse(t *testing.T) {
	token, err := Sign("secret", 42, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	claims, err := Parse("secret", token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", claims.UserID)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	token, err := Sign("secret", 42, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := Parse("other-secret", token); err == nil {
		t.Fatal("expected invalid signature")
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	token, err := Sign("secret", 42, -time.Second)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := Parse("secret", token); err == nil {
		t.Fatal("expected expired token")
	}
}

func TestBearer(t *testing.T) {
	token, err := Bearer("Bearer abc")
	if err != nil {
		t.Fatalf("parse bearer: %v", err)
	}
	if token != "abc" {
		t.Fatalf("expected abc, got %s", token)
	}
	if _, err := Bearer("abc"); err == nil {
		t.Fatal("expected missing bearer token error")
	}
}
