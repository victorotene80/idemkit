package idem

import (
	"context"
	"time"
)

// Store defines the storage-agnostic contract for idempotency.
//
// First principles:
// - Begin must be atomic for a given Key (scope+id).
// - Begin decides NEW vs REPLAY vs CONFLICT.
// - NEW returns a Token used to Commit/Fail safely.
// - Commit/Fail must be conditional on (Key, Token) to prevent races.
// - Store is storage-agnostic (SQL/Redis/etc). No infra dependencies here.
type Store interface {
	Begin(ctx context.Context, req BeginRequest) (BeginResult, error)
	Commit(ctx context.Context, req CommitRequest) error
	Fail(ctx context.Context, req FailRequest) error
}

// BeginRequest is the minimum information required to make an idempotency decision.
type BeginRequest struct {
	Key         Key
	Fingerprint Fingerprint

	// Caller-provided time. Use UTC in your app.
	Now time.Time

	// Optional. If > 0, store may expire in-flight or even finalized records.
	// Policy is store-specific.
	TTL time.Duration
}

// BeginResult is returned by Store.Begin.
type BeginResult struct {
	Decision Decision

	// Token is set only when Decision == DecisionNew.
	// It must be provided to Commit/Fail.
	Token Token

	// Existing is populated for REPLAY/CONFLICT and for debugging/observability.
	Existing StoredResult
}

// Token is an opaque fencing token returned by Begin when DecisionNew.
// It prevents concurrent request A from committing request B.
type Token string

func (t Token) String() string { return string(t) }

// StoredResult represents what the store already knows about a Key.
// Response bytes are opaque (JSON/protobuf/etc).
type StoredResult struct {
	// Stored fingerprint for conflict detection.
	Fingerprint Fingerprint

	// Final indicates a terminal state.
	// If false, record is in-flight.
	Final bool

	// Success indicates terminal success.
	Success bool

	// Response is typically meaningful only if Final && Success.
	Response []byte

	// Optional failure info (store may or may not support persisting it).
	FailureCode string
	FailureMsg  string

	// Optional timestamps (zero time is allowed).
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CommitRequest finalizes a successful attempt and stores the response.
// Must be conditional on (Key, Token). If token doesn't match, return ErrNotOwner.
type CommitRequest struct {
	Key   Key
	Token Token

	Record CommitRecord
	Now    time.Time
}

// FailRequest finalizes a failed attempt.
// Policy (whether failures are replayable) is decided by Manager later.
type FailRequest struct {
	Key   Key
	Token Token

	Record FailRecord
	Now    time.Time
}

// CommitRecord stores a cached success response.
// ResponseBytes is opaque to idemkit.
type CommitRecord struct {
	ResponseBytes []byte

	// Optional metadata for observability/debugging.
	// Determinism tip: prefer stable keys/values.
	Meta map[string]string
}

// FailRecord stores failure details.
// Keep it string-based to avoid ambiguity across languages.
type FailRecord struct {
	Code string
	Msg  string

	Meta map[string]string
}
