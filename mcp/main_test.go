package main

import "testing"

func TestCredentialFromCookieRejectsInvalidCookie(t *testing.T) {
	if _, err := credentialFromCookie("bad-cookie"); err == nil {
		t.Fatal("expected invalid cookie to be rejected")
	}
}
