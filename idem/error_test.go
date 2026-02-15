package idem

import (
	"errors"
	"testing"
)

func TestErrors_IsConflict(t *testing.T) {
	ce := ConflictError{Scope: Scope("payments"), Key: MustKey("payments", "k1")}
	if !errors.Is(ce, ErrConflict) {
		t.Fatalf("expected errors.Is(conflict, ErrConflict)=true")
	}
}

func TestErrors_IsInProgress(t *testing.T) {
	ie := InProgressError{Scope: Scope("payments"), Key: MustKey("payments", "k1")}
	if !errors.Is(ie, ErrInProgress) {
		t.Fatalf("expected errors.Is(inprogress, ErrInProgress)=true")
	}
}

func TestErrors_IsReplayWithFailure(t *testing.T) {
	re := ReplayWithFailureError{Scope: Scope("payments"), Key: MustKey("payments", "k1"), Code: "X", Msg: "Y"}
	if !errors.Is(re, ErrReplayWithFailure) {
		t.Fatalf("expected errors.Is(replayFailure, ErrReplayWithFailure)=true")
	}
}

func TestErrors_IsStoreUnavailable(t *testing.T) {
	base := errors.New("redis down")
	err := StoreUnavailableError{Err: base}
	if !errors.Is(err, ErrStoreUnavailable) {
		t.Fatalf("expected errors.Is(storeUnavailable, ErrStoreUnavailable)=true")
	}
}
