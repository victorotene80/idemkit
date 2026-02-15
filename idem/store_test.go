package idem

import (
	"context"
	"testing"
	"time"
)

func TestStoreInterfaceCompiles(t *testing.T) {
	// This test locks the Store interface shape.
	// If someone changes Store signatures, this will fail to compile.
	var _ Store = (*dummyStore)(nil)

	_ = context.Background()
	_ = time.Now()
}

type dummyStore struct{}

func (d *dummyStore) Begin(ctx context.Context, req BeginRequest) (BeginResult, error) {
	// minimal stub
	return BeginResult{
		Decision: DecisionNew,
		Token:    Token("tok"),
	}, nil
}

func (d *dummyStore) Commit(ctx context.Context, req CommitRequest) error {
	return nil
}

func (d *dummyStore) Fail(ctx context.Context, req FailRequest) error {
	return nil
}
