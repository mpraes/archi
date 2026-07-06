// Package ui provides terminal feedback (spinner) and browser launching.
package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// OpenBrowser opens the given URL in the user's default browser. Errors are returned
// but callers may treat them as non-fatal (server still runs).
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return openBrowserLinux(url)
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

func openBrowserLinux(url string) error {
	var cmds [][]string
	if isWSL() {
		// Prefer WSL-native bridges first when available.
		cmds = append(cmds, []string{"wslview", url}, []string{"explorer.exe", url})
	}
	cmds = append(cmds, []string{"xdg-open", url})

	var errs []string
	for _, c := range cmds {
		if _, err := exec.LookPath(c[0]); err != nil {
			errs = append(errs, c[0]+" not found")
			continue
		}
		if err := exec.Command(c[0], c[1:]...).Start(); err == nil {
			return nil
		} else {
			errs = append(errs, c[0]+": "+err.Error())
		}
	}

	joined := strings.Join(errs, "; ")
	if joined == "" {
		joined = "no browser launcher available"
	}
	return errors.New(joined)
}

func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	b, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(b)), "microsoft")
}

// Spinner prints a minimal animated spinner with a label to stderr (RNF-006).
type Spinner struct {
	label  string
	stop   chan struct{}
	done   chan struct{}
	frames []string
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
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			case <-ticker.C:
				frame := s.frames[i%len(s.frames)]
				fmt.Fprintf(os.Stderr, "\r\033[K%s %s", frame, s.label)
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
