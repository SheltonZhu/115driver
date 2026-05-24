package tools

import "testing"

func TestNormalizeDirectoryListLimitPreservesUnpaginatedDefault(t *testing.T) {
	offset, limit := normalizeDirectoryListPagination(25, 0)
	if offset != 0 {
		t.Fatalf("unexpected offset: %d", offset)
	}
	if limit != 0 {
		t.Fatalf("expected zero limit to request full listing, got %d", limit)
	}
}

func TestNormalizeDirectoryListLimitCapsLargeLimit(t *testing.T) {
	_, limit := normalizeDirectoryListPagination(0, maxDirectoryListLimit+1)
	if limit != maxDirectoryListLimit {
		t.Fatalf("expected max limit %d, got %d", maxDirectoryListLimit, limit)
	}
}
