package idem

import "errors"

var (
	ErrInvalidScope       = errors.New("invalid scope")
	ErrInvalidKey         = errors.New("invalid key")
	ErrInvalidFingerprint = errors.New("invalid fingerprint")
	ErrNotOwner           = errors.New("not owner of idempotency key")
	ErrAlreadyFinal       = errors.New("idempotency key already finalized")
	ErrInFlight           = errors.New("idempotency key is in-flight")
	ErrStoreUnavailable   = errors.New("store unavailable")
	ErrInvalidToken       = errors.New("invalid token")
)
