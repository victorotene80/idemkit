package idem

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidScope        = errors.New("invalid scope")
	ErrInvalidKey          = errors.New("invalid key")
	ErrInvalidFingerprint  = errors.New("invalid fingerprint")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidJSON         = errors.New("invalid JSON")
	ErrConflict            = errors.New("idempotency conflict")
	ErrInProgress          = errors.New("idempotency key is in progress")
	ErrReplayMissingCached = errors.New("replay missing cached response")
	ErrReplayWithFailure   = errors.New("replay is a stored failure")
	ErrStoreUnavailable    = errors.New("store unavailable")
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

func (e ConflictError) Unwrap() error { return ErrConflict }

type InProgressError struct {
	Scope Scope
	Key   Key
}

func (e InProgressError) Error() string {
	return fmt.Sprintf("idempotency in progress: %s", e.Key.String())
}

func (e InProgressError) Unwrap() error { return ErrInProgress }

type ReplayWithFailureError struct {
	Scope Scope
	Key   Key

	Code string
	Msg  string
}

func (e ReplayWithFailureError) Error() string {
	return fmt.Sprintf("idempotency replayed failure: %s (code=%s msg=%s)", e.Key.String(), e.Code, e.Msg)
}

func (e ReplayWithFailureError) Unwrap() error { return ErrReplayWithFailure }

type StoreUnavailableError struct {
	Err error
}

func (e StoreUnavailableError) Error() string {
	if e.Err == nil {
		return ErrStoreUnavailable.Error()
	}
	return fmt.Sprintf("%s: %v", ErrStoreUnavailable.Error(), e.Err)
}

func (e StoreUnavailableError) Unwrap() error {
	if e.Err == nil {
		return ErrStoreUnavailable
	}
	return errors.Join(ErrStoreUnavailable, e.Err)
}

func wrapStoreUnavailable(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrStoreUnavailable) {
		return StoreUnavailableError{Err: err}
	}
	return err
}
