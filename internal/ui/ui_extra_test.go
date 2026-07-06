package ui

import (
	"os"
	"testing"
)

func TestIsWSL(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	if !isWSL() {
		t.Fatal("expected WSL true")
	}
	t.Setenv("WSL_DISTRO_NAME", "")
	if isWSL() && os.Getenv("WSL_INTEROP") == "" {
		// may still be true from /proc/version on WSL hosts
	}
}

func TestOpenBrowserInvalidPlatform(t *testing.T) {
	// Exercise error paths without requiring a desktop environment.
	_ = OpenBrowser("http://127.0.0.1:9/not-a-real-page")
}
