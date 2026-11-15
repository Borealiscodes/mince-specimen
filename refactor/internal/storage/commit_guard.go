// Package storage provides zero-leak, two-phase commit-on-accept persistence
// for Content-Coupled Retrieval (CCR) stores, eliminating orphaned disk records
// when compression candidates fail savings thresholds.
// Zero third-party dependencies (Go standard library only).
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

const (
	// MaxPendingCandidateBytes caps in-memory staging before evaluating compression gain (16 MiB).
	MaxPendingCandidateBytes = 16 * 1024 * 1024
)

var (
	ErrNilBackingStore   = errors.New("storage: backing store cannot be nil")
	ErrCandidateTooLarge = errors.New("storage: candidate exceeds maximum staging budget")
	ErrCandidateRejected = errors.New("storage: compression savings below acceptance threshold; disk write aborted")
	ErrAlreadyFinalized  = errors.New("storage: staged candidate has already been committed or aborted")
)

// BackingStore abstracts persistent CCR persistence (e.g., disk-backed store).
type BackingStore interface {
	Put(id string, raw []byte) error
}

// StagedCandidate holds raw bytes in memory until compression acceptance is confirmed.
type StagedCandidate struct {
	id        string
	raw       []byte
	finalized bool
}

// Stage validates size bounds and computes a deterministic content-coupled ID
// purely in memory without touching persistent disk storage.
func Stage(raw []byte) (*StagedCandidate, error) {
	if len(raw) > MaxPendingCandidateBytes {
		return nil, ErrCandidateTooLarge
	}
	sum := sha256.Sum256(raw)
	id := "ccr_" + hex.EncodeToString(sum[:12])
	cp := make([]byte, len(raw))
	copy(cp, raw)

	return &StagedCandidate{
		id:  id,
		raw: cp,
	}, nil
}

// ID returns the deterministic content-coupled identifier computed during Stage.
func (s *StagedCandidate) ID() string {
	return s.id
}

// CommitIfAccepted persists the staged payload to `dst` if and only if `mincedBytes`
// achieves a strict net byte reduction (including stub overhead).
// On rejection, the staging buffer is zeroed and released with zero disk I/O.
func (s *StagedCandidate) CommitIfAccepted(dst BackingStore, mincedBytes int) (string, error) {
	if s.finalized {
		return "", ErrAlreadyFinalized
	}
	if dst == nil {
		s.abort()
		return "", ErrNilBackingStore
	}
	if mincedBytes < 0 || mincedBytes >= len(s.raw) {
		s.abort()
		return "", ErrCandidateRejected
	}
	if err := dst.Put(s.id, s.raw); err != nil {
		s.abort()
		return "", err
	}
	s.finalized = true
	s.raw = nil // release staging memory after durable write
	return s.id, nil
}

func (s *StagedCandidate) abort() {
	for i := range s.raw {
		s.raw[i] = 0
	}
	s.raw = nil
	s.finalized = true
}
