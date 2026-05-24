package tools

import "testing"

func TestNormalizeDirectoryListLimitDefaultsToBoundedPage(t *testing.T) {
	offset, limit := normalizeDirectoryListPagination(0, 0)
	if offset != 0 {
		t.Fatalf("unexpected offset: %d", offset)
	}
	if limit != defaultDirectoryListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultDirectoryListLimit, limit)
	}
}

func TestNormalizeDirectoryListLimitCapsLargeLimit(t *testing.T) {
	_, limit := normalizeDirectoryListPagination(0, maxDirectoryListLimit+1)
	if limit != maxDirectoryListLimit {
		t.Fatalf("expected max limit %d, got %d", maxDirectoryListLimit, limit)
	}
}
