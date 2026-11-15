// Package storage provides path-traversal-proof, race-safe CCR persistence
// hardening for upstream internal/storage/ccr/store.go.
// Zero third-party dependencies (Go standard library only).
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package storage

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	ccrMagic   = "MCCR1"
	ccrFilePerm = 0o600
	ccrDirPerm  = 0o700
)

var (
	ErrInvalidCCRID    = errors.New("storage: invalid CCR ID (must be 12..64 lowercase hex chars)")
	ErrCorruptCCRFrame = errors.New("storage: corrupt MCCR1 frame or CRC32 mismatch")
)

// NormalizeCCRID strips an optional "ccr_" prefix and validates that the identifier
// consists strictly of 12 to 64 lowercase hexadecimal characters, preventing path traversal.
func NormalizeCCRID(rawID string) (string, error) {
	id := strings.TrimPrefix(rawID, "ccr_")
	if len(id) < 12 || len(id) > 64 {
		return "", ErrInvalidCCRID
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		isDigit := c >= '0' && c <= '9'
		isHex := c >= 'a' && c <= 'f'
		if !isDigit && !isHex {
			return "", ErrInvalidCCRID
		}
	}
	return id, nil
}

// SafeCCRPath resolves a validated CCR ID inside `dir`, failing closed on traversal attempts.
func SafeCCRPath(dir, rawID string) (string, error) {
	id, err := NormalizeCCRID(rawID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ccr_"+id+".gz"), nil
}

// WriteVerifiedAtomic writes an MCCR1 gzip+CRC32 frame using a unique os.CreateTemp file
// so concurrent writers for the same ID never collide on a static .tmp path.
func WriteVerifiedAtomic(dir, rawID string, data []byte) error {
	dstPath, err := SafeCCRPath(dir, rawID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, ccrDirPerm); err != nil {
		return err
	}

	var frame bytes.Buffer
	frame.WriteString(ccrMagic)
	var crc [4]byte
	binary.BigEndian.PutUint32(crc[:], crc32.ChecksumIEEE(data))
	frame.Write(crc[:])

	zw := gzip.NewWriter(&frame)
	if _, err := zw.Write(data); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".ccr-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmpFile.Chmod(ccrFilePerm); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if _, err := tmpFile.Write(frame.Bytes()); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, dstPath); err != nil {
		return err
	}
	cleanup = false
	return nil
}

// ReadVerifiedAtomic reads and CRC32-verifies an MCCR1 frame from a validated path.
func ReadVerifiedAtomic(dir, rawID string) ([]byte, error) {
	srcPath, err := SafeCCRPath(dir, rawID)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}
	if len(raw) < len(ccrMagic)+4 || string(raw[:len(ccrMagic)]) != ccrMagic {
		return nil, ErrCorruptCCRFrame
	}
	wantCRC := binary.BigEndian.Uint32(raw[len(ccrMagic) : len(ccrMagic)+4])
	zr, err := gzip.NewReader(bytes.NewReader(raw[len(ccrMagic)+4:]))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptCCRFrame, err)
	}
	defer zr.Close()

	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptCCRFrame, err)
	}
	if crc32.ChecksumIEEE(out) != wantCRC {
		return nil, ErrCorruptCCRFrame
	}
	return out, nil
}
