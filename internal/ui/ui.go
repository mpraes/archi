// Package ui provides terminal feedback (spinner) and browser launching.
package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// OpenBrowser opens the given URL in the user's default browser. Errors are returned
// but callers may treat them as non-fatal (server still runs).
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// Spinner prints a minimal animated spinner with a label to stderr (RNF-006).
type Spinner struct {
	label  string
	stop   chan struct{}
	done   chan struct{}
	frames []string
	mu     sync.Mutex
}

func NewSpinner(label string) *Spinner {
	return &Spinner{
		label:  label,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func (s *Spinner) Start() {
	go func() {
		defer close(s.done)
		i := 0
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.stop:
				fmt.Printf("\r\033[K")
				return
			case <-ticker.C:
				frame := s.frames[i%len(s.frames)]
				fmt.Printf("\r\033[K%s %s", frame, s.label)
				i++
			}
		}
	}()
}

func (s *Spinner) Stop() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	<-s.done
}

func init() {
	// silence unused import warning if exec path changes
	_ = strings.TrimSpace
}
