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

func TestBuildLSJSONResponseIncludesPaginationMetadata(t *testing.T) {
	resp := buildLSJSONResponse("/", nil, 10, 5)
	if resp["offset"] != int64(10) {
		t.Fatalf("unexpected offset: %v", resp["offset"])
	}
	if resp["limit"] != int64(5) {
		t.Fatalf("unexpected limit: %v", resp["limit"])
	}
	if resp["has_more"] != false {
		t.Fatalf("unexpected has_more: %v", resp["has_more"])
	}
}
