package cmd

import "testing"

func TestRmDirectoryRequiresForceForJSONOutput(t *testing.T) {
	if err := validateDeleteConfirmation(true, true, false); err == nil {
		t.Fatal("expected json directory delete without force to be rejected")
	}
}

func TestRmDirectoryForceSkipsConfirmation(t *testing.T) {
	if err := validateDeleteConfirmation(true, true, true); err != nil {
		t.Fatalf("expected force delete to be allowed: %v", err)
	}
}

func TestRmFileDoesNotRequireForce(t *testing.T) {
	if err := validateDeleteConfirmation(false, true, false); err != nil {
		t.Fatalf("expected file delete to be allowed: %v", err)
	}
}
