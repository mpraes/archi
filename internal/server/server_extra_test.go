package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func newTestRouter(t *testing.T) *httptest.Server {
	t.Helper()
	s := New(sampleSummary(), nil, slog.Default())
	r := chi.NewRouter()
	r.Get("/api/metrics", s.handleMetrics)
	r.Get("/api/modules/{module}", s.handleModule)
	r.Get("/api/ai/enabled", s.handleAIEnabled)
	return httptest.NewServer(r)
}

func TestRouterMetricsAndModule(t *testing.T) {
	ts := newTestRouter(t)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/api/metrics")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("metrics status=%d", res.StatusCode)
	}

	res2, err := http.Get(ts.URL + "/api/modules/alpha")
	if err != nil {
		t.Fatal(err)
	}
	res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("module status=%d", res2.StatusCode)
	}
}

func TestRequestLoggerMiddleware(t *testing.T) {
	logger := slog.Default()
	mw := requestLogger(logger)
	called := false
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("handler not called")
	}
}

func TestSampleSummaryFields(t *testing.T) {
	s := sampleSummary()
	if s.Modules[0].Module != "alpha" {
		t.Fatal("sample summary")
	}
}
