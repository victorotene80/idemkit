package idem

import "testing"

func TestCanonicalizeJSON_DeterministicAcrossFieldOrder(t *testing.T) {
	a := []byte(`{"b":2,"a":1}`)
	b := []byte(`{"a":1,"b":2}`)

	fp1, canon1, err := FingerprintRequest(a, CanonicalizeJSON)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	fp2, canon2, err := FingerprintRequest(b, CanonicalizeJSON)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if string(canon1) != string(canon2) {
		t.Fatalf("expected same canonical bytes:\n%s\n%s", canon1, canon2)
	}
	if fp1 != fp2 {
		t.Fatalf("expected same fingerprint for semantically equal JSON")
	}
}

func TestCanonicalizeJSON_InvalidJSON(t *testing.T) {
	_, err := CanonicalizeJSON([]byte(`{"a":`))
	if err != ErrInvalidJSON {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}
}

func TestCanonicalizeJSON_SameSemanticDifferentOrderSameFingerprint(t *testing.T) {
	a := []byte(`{"b":2,"a":1}`)
	b := []byte(`{"a":1,"b":2}`)

	f1, c1, err := FingerprintRequest(a, CanonicalizeJSON)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	f2, c2, err := FingerprintRequest(b, CanonicalizeJSON)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if string(c1) != string(c2) {
		t.Fatalf("expected same canonical bytes:\n%s\n%s", c1, c2)
	}
	if f1 != f2 {
		t.Fatalf("expected same fingerprint for semantically equal JSON")
	}
}

func TestCanonicalizeJSON_Invalid(t *testing.T) {
	_, err := CanonicalizeJSON([]byte(`{"a":`))
	if err != ErrInvalidJSON {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}
}
