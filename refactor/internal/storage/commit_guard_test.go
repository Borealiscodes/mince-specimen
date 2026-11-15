// Package storage tests two-phase commit-on-accept CCR persistence,
// proving zero disk writes on rejected candidates and bounded RAM staging.
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package storage

import (
	"bytes"
	"errors"
	"testing"
)

type spyBackingStore struct {
	puts map[string][]byte
}

func newSpyStore() *spyBackingStore {
	return &spyBackingStore{puts: make(map[string][]byte)}
}

func (s *spyBackingStore) Put(id string, raw []byte) error {
	cp := make([]byte, len(raw))
	copy(cp, raw)
	s.puts[id] = cp
	return nil
}

func TestStage_EnforcesStagingBudgetAndCopiesInput(t *testing.T) {
	oversized := make([]byte, MaxPendingCandidateBytes+1)
	if _, err := Stage(oversized); !errors.Is(err, ErrCandidateTooLarge) {
		t.Fatalf("expected ErrCandidateTooLarge, got %v", err)
	}

	orig := []byte("original-unmutated-payload")
	cand, err := Stage(orig)
	if err != nil {
		t.Fatalf("Stage failed: %v", err)
	}
	// Mutating caller slice must not alter staged copy
	orig[0] = 'X'
	if cand.raw[0] == 'X' {
		t.Fatal("Stage aliased caller slice instead of defensive copy")
	}
}

func TestCommitIfAccepted_ZeroDiskWritesOnRejectionAndScrubsBuffer(t *testing.T) {
	store := newSpyStore()
	payload := bytes.Repeat([]byte("secret-uncompressible-log-"), 64)

	cand, err := Stage(payload)
	if err != nil {
		t.Fatalf("Stage failed: %v", err)
	}
	backingBuf := cand.raw // keep slice header to verify zero-scrubbing on abort

	// Candidate fails net size reduction (mincedBytes >= len(payload))
	id, err := cand.CommitIfAccepted(store, len(payload))
	if !errors.Is(err, ErrCandidateRejected) {
		t.Fatalf("expected ErrCandidateRejected, got id=%q err=%v", id, err)
	}
	if len(store.puts) != 0 {
		t.Fatalf("expected 0 Put() calls on rejected candidate, got %d", len(store.puts))
	}
	for i, b := range backingBuf {
		if b != 0 {
			t.Fatalf("staging buffer was not zero-scrubbed at byte %d: got 0x%02x", i, b)
		}
	}

	// Second call on finalized candidate fails closed
	if _, err := cand.CommitIfAccepted(store, 10); !errors.Is(err, ErrAlreadyFinalized) {
		t.Fatalf("expected ErrAlreadyFinalized on replay, got %v", err)
	}
}

func TestCommitIfAccepted_PersistsDurableRecordWhenReduced(t *testing.T) {
	store := newSpyStore()
	payload := bytes.Repeat([]byte("compressible-build-trace-line\n"), 100)

	cand, err := Stage(payload)
	if err != nil {
		t.Fatalf("Stage failed: %v", err)
	}

	expectedID := cand.ID()
	gotID, err := cand.CommitIfAccepted(store, 240) // strict reduction: 240 < 3000
	if err != nil {
		t.Fatalf("CommitIfAccepted failed: %v", err)
	}
	if gotID != expectedID {
		t.Fatalf("got ID %q, want %q", gotID, expectedID)
	}
	if len(store.puts) != 1 {
		t.Fatalf("expected exactly 1 Put() call, got %d", len(store.puts))
	}
	if !bytes.Equal(store.puts[gotID], payload) {
		t.Fatal("persisted payload does not match original bytes")
	}
}
