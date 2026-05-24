package server

import (
	"testing"
	"time"
)

func TestNewServerPreservesTransferDefaults(t *testing.T) {
	s := NewServer()
	if s.downloadTimeout != 2*time.Hour {
		t.Fatalf("unexpected default download timeout: %s", s.downloadTimeout)
	}
	if s.urlUploadMaxBytes != 2<<30 {
		t.Fatalf("unexpected default URL upload limit: %d", s.urlUploadMaxBytes)
	}
	if s.downloadMaxBytes != 0 {
		t.Fatalf("expected default download limit to be disabled, got %d", s.downloadMaxBytes)
	}
}
