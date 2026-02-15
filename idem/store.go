package idem

import (
	"context"
	"time"
)

type Store interface {
	Begin(
		ctx context.Context,
		req BeginRequest,
	) (BeginResult, error)

	Commit(
		ctx context.Context,
		req CommitRequest,
	) error

	Fail(
		ctx context.Context,
		req FailRequest,
	) error
}
type Token string

func (t Token) String() string { return string(t) }

type BeginRequest struct {
	Scope       Scope
	Key         Key
	Fingerprint Fingerprint
	Now         time.Time
	TTL         time.Duration
}

type BeginResult struct {
	Decision            Decision
	Token               Token
	Cached              *StoredResult
	ExistingFingerprint Fingerprint
}

type StoredResult struct {
	Fingerprint   Fingerprint
	Final         bool
	Success       bool
	ResponseBytes []byte
	FailureCode   string
	FailureMsg    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CommitRequest struct {
	Scope         Scope
	Key           Key
	Token         Token
	ResponseBytes []byte
	Meta          map[string]string
	Now           time.Time
}

type FailRequest struct {
	Scope Scope
	Key   Key

	Token Token

	Code string
	Msg  string

	Meta map[string]string

	Now time.Time
}
