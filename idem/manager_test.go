package idem

import "testing"

func TestKey_NormalizeAndString(t *testing.T) {
	k, err := NewKey(Scope(" Payments "), "req_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k.Scope().String() != "payments" {
		t.Fatalf("expected normalized scope 'payments', got %q", k.Scope().String())
	}
	if k.String() != "payments:req_123" {
		t.Fatalf("unexpected key string: %q", k.String())
	}
}

func TestKey_Invalid(t *testing.T) {
	_, err := NewKey(Scope(""), "x")
	if err != ErrInvalidScope {
		t.Fatalf("expected ErrInvalidScope, got %v", err)
	}
	_, err = NewKey(Scope("payments"), "")
	if err != ErrInvalidKey {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestFingerprint_SHA256_Deterministic(t *testing.T) {
	a := FingerprintSHA256([]byte("hello"))
	b := FingerprintSHA256([]byte("hello"))
	c := FingerprintSHA256([]byte("world"))

	if a != b {
		t.Fatalf("expected same fingerprint for same input")
	}
	if a == c {
		t.Fatalf("expected different fingerprint for different input")
	}
}
