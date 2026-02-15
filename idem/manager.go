package idem

import (
	"context"
	"errors"
	"time"
)

type Manager struct {
	store Store
	ttl   time.Duration
	now   func() time.Time
}

type ManagerOption func(*Manager)

func WithTTL(ttl time.Duration) ManagerOption {
	return func(m *Manager) { m.ttl = ttl }
}

func WithNow(now func() time.Time) ManagerOption {
	return func(m *Manager) { m.now = now }
}

func NewManager(store Store, opts ...ManagerOption) (*Manager, error) {
	if store == nil {
		return nil, ErrStoreUnavailable
	}
	m := &Manager{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(m)
	}
	return m, nil
}

func MustNewManager(store Store, opts ...ManagerOption) *Manager {
	m, err := NewManager(store, opts...)
	if err != nil {
		panic(err)
	}
	return m
}

func (m *Manager) Do(
	ctx context.Context,
	scope Scope,
	keyID string,
	canonical []byte,
	fn func(context.Context) ([]byte, error),
) (Decision, []byte, error) {

	if fn == nil {
		return 0, nil, errors.New("nil fn")
	}

	key, err := NewKey(scope, keyID)
	if err != nil {
		return 0, nil, err
	}

	fp := FingerprintSHA256(canonical)
	now := m.now().UTC()

	br, err := m.store.Begin(ctx, BeginRequest{
		Scope:       key.Scope(),
		Key:         key,
		Fingerprint: fp,
		Now:         now,
		TTL:         m.ttl,
	})
	if err != nil {
		return 0, nil, err
	}

	switch br.Decision {
	case DecisionReplay:
		// First principles: replay must return cached success response.
		if br.Cached == nil || len(br.Cached.ResponseBytes) == 0 {
			return DecisionReplay, nil, ErrReplayMissingCached
		}
		return DecisionReplay, br.Cached.ResponseBytes, nil

	case DecisionConflict:
		return DecisionConflict, nil, ConflictError{
			Scope:    key.Scope(),
			Key:      key,
			Existing: br.ExistingFingerprint,
			Got:      fp,
		}

	case DecisionNew:
		if br.Token == "" {
			// store bug / contract violation
			return 0, nil, ErrInvalidToken
		}

		// Run business function exactly once (per NEW claim).
		resp, bizErr := fn(ctx)
		if bizErr == nil {
			// Commit success. If commit fails, surface commit error (caller may retry).
			cerr := m.store.Commit(ctx, CommitRequest{
				Scope:         key.Scope(),
				Key:           key,
				Token:         br.Token,
				ResponseBytes: resp,
				Meta:          nil,
				Now:           m.now().UTC(),
			})
			if cerr != nil {
				return DecisionNew, nil, cerr
			}
			return DecisionNew, resp, nil
		}

		_ = m.store.Fail(ctx, FailRequest{
			Scope: key.Scope(),
			Key:   key,
			Token: br.Token,
			Code:  "business_error",
			Msg:   bizErr.Error(),
			Meta:  nil,
			Now:   m.now().UTC(),
		})
		return DecisionNew, nil, bizErr

	default:
		return 0, nil, errors.New("unknown decision")
	}
}
