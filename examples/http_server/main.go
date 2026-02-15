package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/victorotene80/idemkit/idem"
)

type memoryStore struct {
	data map[string]idem.StoredResult
}

func newMemoryStore() *memoryStore {
	return &memoryStore{data: make(map[string]idem.StoredResult)}
}

func (s *memoryStore) Begin(ctx context.Context, req idem.BeginRequest) (idem.BeginResult, error) {
	k := req.Key.String()

	existing, ok := s.data[k]
	if !ok {
		s.data[k] = idem.StoredResult{
			Fingerprint: req.Fingerprint,
			Final:       false,
			Success:     false,
			CreatedAt:   req.Now,
			UpdatedAt:   req.Now,
		}
		return idem.BeginResult{
			Decision: idem.DecisionNew,
			Token:    idem.Token("tok"),
		}, nil
	}

	if existing.Fingerprint != req.Fingerprint {
		return idem.BeginResult{
			Decision:            idem.DecisionConflict,
			ExistingFingerprint: existing.Fingerprint,
		}, nil
	}

	if existing.Final && existing.Success {
		return idem.BeginResult{
			Decision: idem.DecisionReplay,
			Cached:   &existing,
		}, nil
	}

	return idem.BeginResult{
		Decision: idem.DecisionInProgress,
	}, nil
}

func (s *memoryStore) Commit(ctx context.Context, req idem.CommitRequest) error {
	k := req.Key.String()
	s.data[k] = idem.StoredResult{
		Fingerprint:   s.data[k].Fingerprint,
		Final:         true,
		Success:       true,
		ResponseBytes: req.ResponseBytes,
		CreatedAt:     s.data[k].CreatedAt,
		UpdatedAt:     req.Now,
	}
	return nil
}

func (s *memoryStore) Fail(ctx context.Context, req idem.FailRequest) error {
	k := req.Key.String()
	s.data[k] = idem.StoredResult{
		Fingerprint: s.data[k].Fingerprint,
		Final:       true,
		Success:     false,
		FailureCode: req.Code,
		FailureMsg:  req.Msg,
		CreatedAt:   s.data[k].CreatedAt,
		UpdatedAt:   req.Now,
	}
	return nil
}

// ----- HTTP handler -----

func main() {
	store := newMemoryStore()
	manager := idem.MustNewManager(store)

	http.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "missing Idempotency-Key", 400)
			return
		}

		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		decision, resp, err := manager.Do(
			r.Context(),
			idem.Scope("payments"),
			key,
			body,
			func(ctx context.Context) ([]byte, error) {
				// Simulate business logic
				time.Sleep(100 * time.Millisecond)

				out := map[string]string{
					"status": "processed",
				}
				return json.Marshal(out)
			},
		)

		switch decision {
		case idem.DecisionReplay:
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
			return

		case idem.DecisionNew:
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
			return

		case idem.DecisionConflict:
			http.Error(w, err.Error(), http.StatusConflict)
			return

		case idem.DecisionInProgress:
			http.Error(w, "processing", http.StatusConflict)
			return
		}

		if errors.Is(err, idem.ErrStoreUnavailable) {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
