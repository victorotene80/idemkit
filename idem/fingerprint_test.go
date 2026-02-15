package idem

import "testing"

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
