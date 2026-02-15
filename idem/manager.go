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
	canon Canonicalizer
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
	reqBytes []byte,
	fn func(context.Context) ([]byte, error),
) (Decision, []byte, error) {

	if fn == nil {
		return 0, nil, errors.New("nil fn")
	}

	key, err := NewKey(scope, keyID)
	if err != nil {
		return 0, nil, err
	}

	fp, _, err := FingerprintRequest(reqBytes, m.canon)
	if err != nil {
		return 0, nil, err
	}

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
			return 0, nil, ErrInvalidToken
		}

		resp, bizErr := fn(ctx)
		if bizErr == nil {
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

func WithCanonicalizer(c Canonicalizer) ManagerOption {
	return func(m *Manager) { m.canon = c }
}
