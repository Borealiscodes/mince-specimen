// Package wireagent provides deterministic byte-budget ceilings and UTF-8 safe
// truncation for raw CCR retrieval responses (`mince_retrieve`).
// Zero third-party dependencies (Go standard library only).
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package wireagent

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

const (
	// DefaultRetrieveCeilingBytes is the hard maximum payload returned per retrieve call (32 KiB).
	DefaultRetrieveCeilingBytes = 32 * 1024
	// HardMaxCeilingBytes prevents misconfigured budgets from exceeding 128 KiB.
	HardMaxCeilingBytes = 128 * 1024
)

var (
	ErrInvalidCeiling = errors.New("wireagent: retrieval ceiling must be between 1 and 131072 bytes")
	ErrNilReader      = errors.New("wireagent: source reader cannot be nil")
)

// BoundedResult holds a strictly capped retrieval slice along with truncation telemetry.
type BoundedResult struct {
	Data          []byte
	BytesReturned int
	Truncated     bool
}

// ReadBounded reads up to `maxBytes` from `r`, clamps the ceiling to `HardMaxCeilingBytes`,
// avoids splitting multi-byte UTF-8 runes at the cutoff, and reports whether excess data was truncated.
func ReadBounded(r io.Reader, maxBytes int) (BoundedResult, error) {
	if r == nil {
		return BoundedResult{}, ErrNilReader
	}
	if maxBytes <= 0 || maxBytes > HardMaxCeilingBytes {
		return BoundedResult{}, ErrInvalidCeiling
	}

	// Read at most maxBytes + 1 to detect whether additional unread bytes remain.
	limited := io.LimitReader(r, int64(maxBytes)+1)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(limited); err != nil {
		return BoundedResult{}, err
	}

	raw := buf.Bytes()
	if len(raw) <= maxBytes {
		return BoundedResult{
			Data:          raw,
			BytesReturned: len(raw),
			Truncated:     false,
		}, nil
	}

	// Trim down to valid UTF-8 rune boundary within maxBytes.
	trimmed := trimValidUTF8(raw[:maxBytes])
	return BoundedResult{
		Data:          trimmed,
		BytesReturned: len(trimmed),
		Truncated:     true,
	}, nil
}

// trimValidUTF8 steps backward at most 3 bytes so multi-byte UTF-8 characters are never split mid-rune.
func trimValidUTF8(b []byte) []byte {
	for len(b) > 0 && !utf8.Valid(b) {
		// Only trim trailing incomplete rune sequences (up to 3 continuation bytes).
		if len(b) >= 4 && utf8.Valid(b[:len(b)-3]) {
			return b[:len(b)-3]
		}
		b = b[:len(b)-1]
	}
	return b
}
