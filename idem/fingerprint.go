package idem

import (
	"crypto/sha256"
	"encoding/hex"
)

type Canonicalizer func(reqBytes []byte) ([]byte, error)

type Fingerprint [32]byte

func (f Fingerprint) Hex() string { return hex.EncodeToString(f[:]) }

func (f Fingerprint) IsZero() bool {
	var z Fingerprint
	return f == z
}

func FingerprintSHA256(canonical []byte) Fingerprint {
	sum := sha256.Sum256(canonical)
	return Fingerprint(sum)
}

func FingerprintFromHex(h string) (Fingerprint, error) {
	b, err := hex.DecodeString(h)
	if err != nil || len(b) != 32 {
		return Fingerprint{}, ErrInvalidFingerprint
	}
	var f Fingerprint
	copy(f[:], b)
	return f, nil
}

func FingerprintRequest(reqBytes []byte, canon Canonicalizer) (Fingerprint, []byte, error) {
	if canon != nil {
		c, err := canon(reqBytes)
		if err != nil {
			return Fingerprint{}, nil, err
		}
		return FingerprintSHA256(c), c, nil
	}
	return FingerprintSHA256(reqBytes), reqBytes, nil
}
