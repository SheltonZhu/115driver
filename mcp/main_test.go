package main

import (
	"testing"
	"time"
)

func TestCredentialFromCookieRejectsInvalidCookie(t *testing.T) {
	if _, err := credentialFromCookie("bad-cookie"); err == nil {
		t.Fatal("expected invalid cookie to be rejected")
	}
}

func TestValidateOptionsRejectsNegativeTransferLimits(t *testing.T) {
	if err := validateOptions(-1, 0, time.Hour); err == nil {
		t.Fatal("expected negative URL upload limit to be rejected")
	}
	if err := validateOptions(0, -1, time.Hour); err == nil {
		t.Fatal("expected negative download limit to be rejected")
	}
}

func TestValidateOptionsRejectsNegativeTimeout(t *testing.T) {
	if err := validateOptions(0, 0, -time.Second); err == nil {
		t.Fatal("expected negative timeout to be rejected")
	}
}

func TestValidateOptionsAllowsZeroToDisableLimits(t *testing.T) {
	if err := validateOptions(0, 0, 0); err != nil {
		t.Fatalf("expected zero values to be accepted: %v", err)
	}
}
