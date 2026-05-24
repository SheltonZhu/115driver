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

func TestBuildLSTextPaginationNoticeShowsNextOffsetWhenPageIsFull(t *testing.T) {
	got := buildLSTextPaginationNotice(100, 200, 100)
	want := "Showing 100 entries. Use --offset 300 to continue.\n"
	if got != want {
		t.Fatalf("unexpected notice: got %q want %q", got, want)
	}
}

func TestBuildLSTextPaginationNoticeEmptyWhenPageIsNotFull(t *testing.T) {
	if got := buildLSTextPaginationNotice(99, 0, 100); got != "" {
		t.Fatalf("expected no notice, got %q", got)
	}
}
