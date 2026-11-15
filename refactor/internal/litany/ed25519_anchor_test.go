// Package litany tests asymmetric Ed25519 epoch anchoring over SHA-256 hash-chain heads,
// verifying canonical 84-byte payload binding and fail-closed tamper detection.
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package litany

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"testing"
)

func TestEpochCheckpoint_SignAndVerifyRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	chainHead := sha256.Sum256([]byte("ltny-v3-block-0001"))
	cp, err := SignEpoch(priv, 1, 1794736800, chainHead)
	if err != nil {
		t.Fatalf("SignEpoch failed: %v", err)
	}

	// Verify deterministic 84-byte canonical header layout
	payload := cp.CanonicalPayload()
	if !bytes.Equal(payload[0:4], []byte("LTNY")) {
		t.Fatalf("expected domain prefix LTNY, got %q", payload[0:4])
	}
	if cp.SignerID != sha256.Sum256(pub) {
		t.Fatal("SignerID does not match SHA-256(pub)")
	}

	if err := cp.VerifyEpoch(pub, chainHead); err != nil {
		t.Fatalf("VerifyEpoch failed on valid checkpoint: %v", err)
	}
}

func TestEpochCheckpoint_KeySizeAndSignerMismatchGuards(t *testing.T) {
	chainHead := sha256.Sum256([]byte("ltny-v3-block-0002"))

	// Reject malformed private key length
	if _, err := SignEpoch(ed25519.PrivateKey("short"), 1, 1000, chainHead); !errors.Is(err, ErrInvalidPrivateKey) {
		t.Fatalf("expected ErrInvalidPrivateKey, got %v", err)
	}

	pub1, priv1, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen 1 failed: %v", err)
	}
	pub2, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen 2 failed: %v", err)
	}

	cp, err := SignEpoch(priv1, 7, 1794736900, chainHead)
	if err != nil {
		t.Fatalf("SignEpoch failed: %v", err)
	}

	// Reject malformed public key length
	if err := cp.VerifyEpoch(ed25519.PublicKey("short"), chainHead); !errors.Is(err, ErrInvalidPublicKey) {
		t.Fatalf("expected ErrInvalidPublicKey for short key, got %v", err)
	}

	// Reject valid-length wrong signer key
	if err := cp.VerifyEpoch(pub2, chainHead); !errors.Is(err, ErrInvalidPublicKey) {
		t.Fatalf("expected ErrInvalidPublicKey for wrong signer, got %v", err)
	}

	// Sanity check original key still passes
	if err := cp.VerifyEpoch(pub1, chainHead); err != nil {
		t.Fatalf("expected valid key to verify, got %v", err)
	}
}

func TestEpochCheckpoint_DetectsTamperingAndReplay(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen failed: %v", err)
	}

	headA := sha256.Sum256([]byte("ltny-v3-chain-head-A"))
	headB := sha256.Sum256([]byte("ltny-v3-chain-head-B"))

	cp, err := SignEpoch(priv, 10, 1794737000, headA)
	if err != nil {
		t.Fatalf("SignEpoch failed: %v", err)
	}

	// 1. Chain head substitution fails closed with ErrChainHeadMismatch
	if err := cp.VerifyEpoch(pub, headB); !errors.Is(err, ErrChainHeadMismatch) {
		t.Fatalf("expected ErrChainHeadMismatch, got %v", err)
	}

	// 2. Epoch sequence replay/rollback tampering invalidates signature
	tamperedEpoch := *cp
	tamperedEpoch.EpochSequence = 11
	if err := tamperedEpoch.VerifyEpoch(pub, headA); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature on epoch mutation, got %v", err)
	}

	// 3. Timestamp mutation invalidates signature
	tamperedTime := *cp
	tamperedTime.TimestampUnix++
	if err := tamperedTime.VerifyEpoch(pub, headA); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature on timestamp mutation, got %v", err)
	}

	// 4. Single-bit signature flip invalidates verification
	tamperedSig := *cp
	tamperedSig.Signature[0] ^= 0x01
	if err := tamperedSig.VerifyEpoch(pub, headA); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature on signature bit-flip, got %v", err)
	}
}
