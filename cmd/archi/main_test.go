package main

import "testing"

func TestVersionDefault(t *testing.T) {
	if version == "" {
		t.Fatal("version should not be empty")
	}
}
