package ui

import (
	"testing"
)

func TestIsWSL(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	t.Setenv("WSL_INTEROP", "")
	if !isWSL() {
		t.Fatal("expected WSL true when WSL_DISTRO_NAME is set")
	}
}

func TestOpenBrowserInvalidPlatform(t *testing.T) {
	_ = OpenBrowser("http://127.0.0.1:9/not-a-real-page")
}
