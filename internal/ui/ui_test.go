package ui

import (
	"testing"
	"time"
)

func TestSpinnerLifecycle(t *testing.T) {
	s := NewSpinner("test")
	s.Start()
	time.Sleep(120 * time.Millisecond)
	s.Stop()
}

func TestOpenBrowserUnsupported(t *testing.T) {
	// OpenBrowser on linux may succeed or fail depending on environment;
	// ensure it returns without panic for invalid URL on current platform.
	_ = OpenBrowser("http://127.0.0.1:1")
}
