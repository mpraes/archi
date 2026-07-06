package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := buildBinary(); err != nil {
		fmt.Fprintf(os.Stderr, "build binary: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func binPath(t *testing.T) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "archi")
}

func buildBinary() error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(root, "archi"), "./cmd/archi")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

func TestBinaryExportJSON(t *testing.T) {
	root := fixtureMiniGo(t)
	cmd := exec.Command(binPath(t), "export", root, "--format", "json", "--lang", "go")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v stderr=%s", err, stderr.String())
	}
	var summary struct {
		ProjectName string `json:"projectName"`
		ModuleCount int    `json:"moduleCount"`
		Modules     []any  `json:"modules"`
	}
	if err := json.Unmarshal(out.Bytes(), &summary); err != nil {
		t.Fatalf("json: %v body=%s", err, out.String())
	}
	if summary.ModuleCount < 1 {
		t.Fatalf("moduleCount = %d", summary.ModuleCount)
	}
}

func TestBinaryCheckExitCodes(t *testing.T) {
	root := fixtureMiniGo(t)
	fail := exec.Command(binPath(t), "check", "--max-distance", "0.5", "--lang", "go", root)
	if err := fail.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			t.Fatalf("unexpected err: %v", err)
		}
	} else {
		t.Fatal("expected check failure")
	}
	ok := exec.Command(binPath(t), "check", "--max-distance", "1.0", "--lang", "go", root)
	if err := ok.Run(); err != nil {
		t.Fatalf("expected success: %v", err)
	}
}

func TestBinaryServerAPI(t *testing.T) {
	root := fixtureMiniGo(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, binPath(t), root, "--no-browser", "--port", "0", "--lang", "go")
	var stderr syncBuffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cancel()
		_ = cmd.Wait()
	}()

	addr, err := waitForServerAddr(&stderr, 10*time.Second)
	if err != nil {
		t.Fatalf("server addr: %v stderr=%s", err, stderr.String())
	}
	base := "http://" + addr
	resp, err := http.Get(base + "/api/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("metrics status = %d", resp.StatusCode)
	}
	resp2, err := http.Get(base + "/api/ai/enabled")
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("ai enabled status = %d", resp2.StatusCode)
	}
}

func waitForServerAddr(stderr *syncBuffer, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s := stderr.String()
		if i := strings.Index(s, "http://"); i >= 0 {
			rest := s[i+len("http://"):]
			if j := strings.IndexAny(rest, " \n"); j >= 0 {
				return rest[:j], nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout")
}

// syncBuffer is a mutex-protected buffer safe for concurrent exec writes and test reads.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func fixtureMiniGo(t *testing.T) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "testdata", "mini-go")
}
