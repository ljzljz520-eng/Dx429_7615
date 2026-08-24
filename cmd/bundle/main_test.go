package main

import "testing"

func TestVersionString(t *testing.T) {
	if versionString() == "" {
		t.Fatal("empty version")
	}
}
