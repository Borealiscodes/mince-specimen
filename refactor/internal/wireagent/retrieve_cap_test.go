// Package wireagent tests deterministic byte-budget ceilings and UTF-8 safe
// boundary alignment for bounded CCR retrieval responses.
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package wireagent

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestReadBounded_GuardValidation(t *testing.T) {
	if _, err := ReadBounded(nil, DefaultRetrieveCeilingBytes); !errors.Is(err, ErrNilReader) {
		t.Fatalf("expected ErrNilReader, got %v", err)
	}

	r := strings.NewReader("hello")
	for _, badCeiling := range []int{0, -1, HardMaxCeilingBytes + 1} {
		if _, err := ReadBounded(r, badCeiling); !errors.Is(err, ErrInvalidCeiling) {
			t.Fatalf("ceiling %d: expected ErrInvalidCeiling, got %v", badCeiling, err)
		}
	}
}

func TestReadBounded_ExactFitAndBelowCeilingNotTruncated(t *testing.T) {
	payload := bytes.Repeat([]byte("A"), DefaultRetrieveCeilingBytes)
	res, err := ReadBounded(bytes.NewReader(payload), DefaultRetrieveCeilingBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Truncated {
		t.Fatalf("expected Truncated=false for exact ceiling match (%d bytes)", len(payload))
	}
	if res.BytesReturned != DefaultRetrieveCeilingBytes || len(res.Data) != DefaultRetrieveCeilingBytes {
		t.Fatalf("got %d bytes, want %d", res.BytesReturned, DefaultRetrieveCeilingBytes)
	}
}

func TestReadBounded_ClampsOversizedPayloadAndPreservesUTF8(t *testing.T) {
	// Fill > 32 KiB with 4-byte UTF-8 emojis ("🚀" = 4 bytes each).
	// Using a ceiling of 32767 (DefaultRetrieveCeilingBytes - 1) forces a cut 3 bytes into a 4-byte rune!
	const oddCeiling = DefaultRetrieveCeilingBytes - 1 // 32767 is not divisible by 4 (32767 % 4 == 3)
	oversized := bytes.Repeat([]byte("🚀"), 10_000)   // 40,000 bytes

	res, err := ReadBounded(bytes.NewReader(oversized), oddCeiling)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Truncated {
		t.Fatal("expected Truncated=true for oversized input")
	}
	if res.BytesReturned > oddCeiling || len(res.Data) != res.BytesReturned {
		t.Fatalf("returned %d bytes, exceeds ceiling %d", res.BytesReturned, oddCeiling)
	}
	if !utf8.Valid(res.Data) {
		t.Fatal("truncated output failed utf8.Valid boundary check")
	}
	// Nearest 4-byte boundary <= 32767 is 32764 (8191 rockets * 4 bytes)
	if res.BytesReturned != 32764 {
		t.Fatalf("expected 32764 aligned UTF-8 bytes, got %d", res.BytesReturned)
	}
}

func TestReadBounded_MultiByteCutoffOffsets(t *testing.T) {
	// Test cutting at 1, 2, and 3 bytes into a 4-byte UTF-8 sequence ("🧪").
	runeBytes := []byte("🧪") // 4 bytes
	payload := append(append([]byte("OK"), runeBytes...), []byte("EXTRA")...)

	for offset := 1; offset <= 3; offset++ {
		ceiling := 2 + offset // cuts mid-rune after "OK"
		res, err := ReadBounded(bytes.NewReader(payload), ceiling)
		if err != nil {
			t.Fatalf("offset %d: unexpected error: %v", offset, err)
		}
		if !res.Truncated {
			t.Fatalf("offset %d: expected Truncated=true", offset)
		}
		if !utf8.Valid(res.Data) {
			t.Fatalf("offset %d: invalid UTF-8 output %q", offset, res.Data)
		}
		if string(res.Data) != "OK" {
			t.Fatalf("offset %d: expected clean trim back to %q, got %q", offset, "OK", string(res.Data))
		}
	}
}
