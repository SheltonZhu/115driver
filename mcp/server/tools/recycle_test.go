package tools

import "testing"

func TestNormalizeRecyclePaginationDefaultsAndCaps(t *testing.T) {
	offset, limit := normalizeRecyclePagination(-1, 0)
	if offset != 0 {
		t.Fatalf("unexpected offset: %d", offset)
	}
	if limit != defaultRecycleLimit {
		t.Fatalf("expected default limit %d, got %d", defaultRecycleLimit, limit)
	}

	_, limit = normalizeRecyclePagination(0, maxRecycleLimit+1)
	if limit != maxRecycleLimit {
		t.Fatalf("expected max limit %d, got %d", maxRecycleLimit, limit)
	}
}
