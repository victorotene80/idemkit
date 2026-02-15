package idem

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidScope        = errors.New("invalid scope")
	ErrInvalidKey          = errors.New("invalid key")
	ErrInvalidFingerprint  = errors.New("invalid fingerprint")
	ErrNotOwner            = errors.New("not owner of idempotency key")
	ErrAlreadyFinal        = errors.New("idempotency key already finalized")
	ErrInFlight            = errors.New("idempotency key is in-flight")
	ErrStoreUnavailable    = errors.New("store unavailable")
	ErrInvalidToken        = errors.New("invalid token")
	ErrConflict            = errors.New("idempotency conflict")
	ErrReplayMissingCached = errors.New("replay missing cached response")
	ErrInvalidJSON         = errors.New("invalid JSON")
	ErrInProgress          = errors.New("idempotency key is in progress")
)

type ConflictError struct {
	Scope    Scope
	Key      Key
	Existing Fingerprint
	Got      Fingerprint
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("idempotency conflict: %s (existing=%s got=%s)",
		e.Key.String(), e.Existing.Hex(), e.Got.Hex())
}

type InProgressError struct {
	Scope Scope
	Key   Key
}

func (e InProgressError) Error() string {
	return fmt.Sprintf("idempotency in progress: %s", e.Key.String())
}

func (e InProgressError) Unwrap() error { return ErrInProgress }

/*type ConflictError struct {
	Scope    Scope
	Key      Key
	Existing Fingerprint
	Got      Fingerprint
}

func (e ConflictError) Error() string {
	return fmt.Sprintf(
		"%v: scope=%s key=%s existing_fp=%s got_fp=%s",
		ErrConflict,
		e.Scope.String(),
		e.Key.String(),
		e.Existing.Hex(),
		e.Got.Hex(),
	)
}

func (e ConflictError) Unwrap() error { return ErrConflict }
*/
