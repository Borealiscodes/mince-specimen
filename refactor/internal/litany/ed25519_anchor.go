// Package litany provides asymmetric Ed25519 epoch anchoring over SHA-256 hash-chain heads.
// Zero third-party dependencies (Go standard library only).
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package litany

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidPublicKey   = errors.New("litany: invalid ed25519 public key length")
	ErrInvalidPrivateKey  = errors.New("litany: invalid ed25519 private key length")
	ErrInvalidSignature   = errors.New("litany: ed25519 checkpoint signature verification failed")
	ErrTruncatedFrame     = errors.New("litany: checkpoint frame truncated")
	ErrChainHeadMismatch  = errors.New("litany: checkpoint chain head mismatch")
)

// EpochCheckpoint binds a 32-byte LTNY v3 SHA-256 chain head to an Ed25519 signature.
type EpochCheckpoint struct {
	EpochSequence uint64
	TimestampUnix uint64
	ChainHead     [32]byte
	SignerID      [32]byte // SHA-256(Ed25519_PublicKey)
	Signature     [64]byte
}

// CanonicalPayload serializes the deterministic 84-byte signed message:
// [domain: 4B ("LTNY") | epoch: 8B BE | timestamp: 8B BE | chain_head: 32B | signer_id: 32B]
func (c *EpochCheckpoint) CanonicalPayload() [84]byte {
	var buf [84]byte
	copy(buf[0:4], "LTNY")
	binary.BigEndian.PutUint64(buf[4:12], c.EpochSequence)
	binary.BigEndian.PutUint64(buf[12:20], c.TimestampUnix)
	copy(buf[20:52], c.ChainHead[:])
	copy(buf[52:84], c.SignerID[:])
	return buf
}

// SignEpoch creates an Ed25519-signed checkpoint over a 32-byte LTNY chain head.
func SignEpoch(priv ed25519.PrivateKey, epochSeq, nowUnix uint64, chainHead [32]byte) (*EpochCheckpoint, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, ErrInvalidPrivateKey
	}
	pub := priv.Public().(ed25519.PublicKey)
	signerID := sha256.Sum256(pub)

	cp := &EpochCheckpoint{
		EpochSequence: epochSeq,
		TimestampUnix: nowUnix,
		ChainHead:     chainHead,
		SignerID:      signerID,
	}
	payload := cp.CanonicalPayload()
	sig := ed25519.Sign(priv, payload[:])
	copy(cp.Signature[:], sig)
	return cp, nil
}

// VerifyEpoch validates the signer fingerprint, expected chain head, and Ed25519 signature.
func (c *EpochCheckpoint) VerifyEpoch(pub ed25519.PublicKey, expectedChainHead [32]byte) error {
	if len(pub) != ed25519.PublicKeySize {
		return ErrInvalidPublicKey
	}
	if sha256.Sum256(pub) != c.SignerID {
		return ErrInvalidPublicKey
	}
	if c.ChainHead != expectedChainHead {
		return ErrChainHeadMismatch
	}
	payload := c.CanonicalPayload()
	if !ed25519.Verify(pub, payload[:], c.Signature[:]) {
		return ErrInvalidSignature
	}
	return nil
}
