package storage

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

func TestNormalizeCCRID_BlocksPathTraversal(t *testing.T) {
	badIDs := []string{
		"",
		"short123",
		"../../etc/passwd",
		"ccr_../../secret",
		"abcdef123456/../../x",
		"ABCDEF123456", // uppercase rejected
		"abcdef12345g", // non-hex rejected
	}
	for _, bad := range badIDs {
		if _, err := SafeCCRPath("/tmp/ccr", bad); !errors.Is(err, ErrInvalidCCRID) {
			t.Fatalf("expected ErrInvalidCCRID for %q, got %v", bad, err)
		}
	}

	got, err := NormalizeCCRID("ccr_ab12cd34ef56")
	if err != nil || got != "ab12cd34ef56" {
		t.Fatalf("valid ID failed: got=%q err=%v", got, err)
	}
}

func TestWriteVerifiedAtomic_ConcurrentSameIDRaceSafe(t *testing.T) {
	dir := t.TempDir()
	const id = "ccr_0123456789abcdef"
	payload := bytes.Repeat([]byte("concurrent-ccr-write-verification\n"), 200)

	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := WriteVerifiedAtomic(dir, id, payload); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent WriteVerifiedAtomic failed: %v", err)
	}

	got, err := ReadVerifiedAtomic(dir, id)
	if err != nil {
		t.Fatalf("ReadVerifiedAtomic failed: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatal("round-trip payload mismatch")
	}
}
