package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/mpraes/archi/internal/ai"
	"github.com/mpraes/archi/internal/model"
)

func sampleSummary() model.Summary {
	return model.Summary{
		ProjectName: "demo",
		ModuleCount: 1,
		Modules: []model.ModuleMetrics{{
			Module: "alpha", Path: "alpha", Language: "go",
			Afferent: 0, Efferent: 1, Instability: 1, Distance: 0.2,
		}},
		Hotspots:    []string{},
		Connascence: []model.Connascence{},
	}
}

func withURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestHandleMetrics(t *testing.T) {
	s := New(sampleSummary(), nil, slog.Default())
	rec := httptest.NewRecorder()
	s.handleMetrics(rec, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	var summary model.Summary
	if err := json.NewDecoder(rec.Body).Decode(&summary); err != nil {
		t.Fatal(err)
	}
	if summary.ProjectName != "demo" {
		t.Fatalf("project = %q", summary.ProjectName)
	}
}

func TestHandleModule(t *testing.T) {
	s := New(sampleSummary(), nil, slog.Default())
	rec := httptest.NewRecorder()
	req := withURLParam(httptest.NewRequest(http.MethodGet, "/api/modules/alpha", nil), "module", "alpha")
	s.handleModule(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	rec2 := httptest.NewRecorder()
	req2 := withURLParam(httptest.NewRequest(http.MethodGet, "/api/modules/missing", nil), "module", "missing")
	s.handleModule(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec2.Code)
	}
}

func TestHandleAIEnabled(t *testing.T) {
	off := New(sampleSummary(), nil, slog.Default())
	rec := httptest.NewRecorder()
	off.handleAIEnabled(rec, httptest.NewRequest(http.MethodGet, "/api/ai/enabled", nil))
	var body map[string]bool
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["enabled"] {
		t.Fatal("AI should be disabled")
	}
	on := New(sampleSummary(), &ai.Config{APIKey: "k"}, slog.Default())
	rec2 := httptest.NewRecorder()
	on.handleAIEnabled(rec2, httptest.NewRequest(http.MethodGet, "/api/ai/enabled", nil))
	_ = json.NewDecoder(rec2.Body).Decode(&body)
	if !body["enabled"] {
		t.Fatal("AI should be enabled")
	}
}

func TestHandleAIInsightsDisabled(t *testing.T) {
	s := New(sampleSummary(), nil, slog.Default())
	rec := httptest.NewRecorder()
	s.handleAIInsights(rec, httptest.NewRequest(http.MethodGet, "/api/ai/insights", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWriteSSEAndJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, map[string]string{"ok": "1"})
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatal("content-type")
	}
	rec2 := httptest.NewRecorder()
	writeSSE(rec2, "chunk", "hello")
	body := rec2.Body.String()
	if !stringsContains(body, "event: chunk") {
		t.Fatalf("sse = %q", body)
	}
}

func TestServerRunShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := New(sampleSummary(), nil, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx, ln) }()
	time.Sleep(150 * time.Millisecond)
	resp, err := http.Get("http://" + ln.Addr().String() + "/api/metrics")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for shutdown")
	}
}

func stringsContains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
