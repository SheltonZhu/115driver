package cmd

import "testing"

func TestNormalizeLSPageDefaultsToBoundedLimit(t *testing.T) {
	offset, limit := normalizeLSPage(0, 0)
	if offset != 0 {
		t.Fatalf("unexpected offset: %d", offset)
	}
	if limit != defaultLSLimit {
		t.Fatalf("expected default limit %d, got %d", defaultLSLimit, limit)
	}
}

func TestNormalizeLSPageCapsLargeLimit(t *testing.T) {
	_, limit := normalizeLSPage(0, maxLSLimit+1)
	if limit != maxLSLimit {
		t.Fatalf("expected max limit %d, got %d", maxLSLimit, limit)
	}
}
