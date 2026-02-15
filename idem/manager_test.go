package idem

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memStore struct {
	beginRes BeginResult
	beginErr error

	commitErr error
	failErr   error

	beginCalls  int
	commitCalls int
	failCalls   int

	lastCommit CommitRequest
	lastFail   FailRequest
}

func (s *memStore) Begin(ctx context.Context, req BeginRequest) (BeginResult, error) {
	s.beginCalls++
	return s.beginRes, s.beginErr
}

func (s *memStore) Commit(ctx context.Context, req CommitRequest) error {
	s.commitCalls++
	s.lastCommit = req
	return s.commitErr
}

func (s *memStore) Fail(ctx context.Context, req FailRequest) error {
	s.failCalls++
	s.lastFail = req
	return s.failErr
}

func TestManager_Do_REPLAY_ReturnsCached(t *testing.T) {
	store := &memStore{
		beginRes: BeginResult{
			Decision: DecisionReplay,
			Cached: &StoredResult{
				Final:         true,
				Success:       true,
				ResponseBytes: []byte(`{"ok":true}`),
			},
		},
	}

	m := MustNewManager(store, WithNow(func() time.Time { return time.Unix(1, 0).UTC() }))

	called := 0
	dec, resp, err := m.Do(context.Background(), Scope("payments"), "k1", []byte("canon"), func(ctx context.Context) ([]byte, error) {
		called++
		return nil, nil
	})

	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dec != DecisionReplay {
		t.Fatalf("expected REPLAY, got %v", dec)
	}
	if called != 0 {
		t.Fatalf("expected fn not called, got %d", called)
	}
	if string(resp) != `{"ok":true}` {
		t.Fatalf("unexpected cached resp: %s", string(resp))
	}
}

func TestManager_Do_CONFLICT_ReturnsConflictError(t *testing.T) {
	existing := FingerprintSHA256([]byte("old"))

	store := &memStore{
		beginRes: BeginResult{
			Decision:            DecisionConflict,
			ExistingFingerprint: existing,
		},
	}

	m := MustNewManager(store, WithNow(func() time.Time { return time.Unix(1, 0).UTC() }))

	dec, _, err := m.Do(context.Background(), Scope("payments"), "k1", []byte("new"), func(ctx context.Context) ([]byte, error) {
		return []byte("ok"), nil
	})

	if dec != DecisionConflict {
		t.Fatalf("expected CONFLICT, got %v", dec)
	}

	var ce ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("expected ConflictError, got %T: %v", err, err)
	}
	if ce.Existing != existing {
		t.Fatalf("expected existing fingerprint to match")
	}
}

func TestManager_Do_NEW_CommitsOnSuccess(t *testing.T) {
	store := &memStore{
		beginRes: BeginResult{
			Decision: DecisionNew,
			Token:    Token("tok"),
		},
	}

	m := MustNewManager(store, WithNow(func() time.Time { return time.Unix(1, 0).UTC() }))

	called := 0
	dec, resp, err := m.Do(context.Background(), Scope("payments"), "k1", []byte("canon"), func(ctx context.Context) ([]byte, error) {
		called++
		return []byte("ok"), nil
	})

	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dec != DecisionNew {
		t.Fatalf("expected NEW, got %v", dec)
	}
	if called != 1 {
		t.Fatalf("expected fn called once, got %d", called)
	}
	if store.commitCalls != 1 {
		t.Fatalf("expected commit called once, got %d", store.commitCalls)
	}
	if string(resp) != "ok" {
		t.Fatalf("unexpected resp: %s", string(resp))
	}
	if store.lastCommit.Token != "tok" {
		t.Fatalf("expected commit token tok, got %q", store.lastCommit.Token)
	}
}

func TestManager_Do_NEW_FailsOnBusinessError(t *testing.T) {
	store := &memStore{
		beginRes: BeginResult{
			Decision: DecisionNew,
			Token:    Token("tok"),
		},
	}

	m := MustNewManager(store, WithNow(func() time.Time { return time.Unix(1, 0).UTC() }))

	bizErr := errors.New("boom")

	dec, _, err := m.Do(context.Background(), Scope("payments"), "k1", []byte("canon"), func(ctx context.Context) ([]byte, error) {
		return nil, bizErr
	})

	if dec != DecisionNew {
		t.Fatalf("expected NEW, got %v", dec)
	}
	if !errors.Is(err, bizErr) {
		t.Fatalf("expected business error, got %v", err)
	}
	if store.failCalls != 1 {
		t.Fatalf("expected fail called once, got %d", store.failCalls)
	}
	if store.commitCalls != 0 {
		t.Fatalf("expected commit not called, got %d", store.commitCalls)
	}
}

func TestManager_Do_INPROGRESS_DoesNotExecute(t *testing.T) {
	store := &memStore{
		beginRes: BeginResult{
			Decision: DecisionInProgress,
		},
	}

	m := MustNewManager(store, WithNow(func() time.Time { return time.Unix(1, 0).UTC() }))

	called := 0
	dec, _, err := m.Do(context.Background(), Scope("payments"), "k1", []byte("canon"), func(ctx context.Context) ([]byte, error) {
		called++
		return []byte("ok"), nil
	})

	if dec != DecisionInProgress {
		t.Fatalf("expected IN_PROGRESS, got %v", dec)
	}
	if called != 0 {
		t.Fatalf("expected fn not called, got %d", called)
	}
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrInProgress) {
		t.Fatalf("expected ErrInProgress, got %v", err)
	}
	if store.commitCalls != 0 {
		t.Fatalf("expected commit not called, got %d", store.commitCalls)
	}
	if store.failCalls != 0 {
		t.Fatalf("expected fail not called, got %d", store.failCalls)
	}
}

func TestManager_Do_REPLAY_WithFailure_ReturnsTypedError(t *testing.T) {
	store := &memStore{
		beginRes: BeginResult{
			Decision: DecisionReplay,
			Cached: &StoredResult{
				Final:       true,
				Success:     false,
				FailureCode: "DECLINED",
				FailureMsg:  "insufficient funds",
			},
		},
	}

	m := MustNewManager(store)

	dec, _, err := m.Do(context.Background(), Scope("payments"), "k1", []byte(`{"x":1}`), func(ctx context.Context) ([]byte, error) {
		return []byte("should_not_run"), nil
	})

	if dec != DecisionReplay {
		t.Fatalf("expected REPLAY, got %v", dec)
	}
	if !errors.Is(err, ErrReplayWithFailure) {
		t.Fatalf("expected ErrReplayWithFailure, got %v", err)
	}
}
