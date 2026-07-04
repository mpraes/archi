// Package server serves the embedded web UI and the JSON metrics API.
// It is 100% offline-capable (RNF-007); AI insights are an optional streamed
// endpoint that never blocks initial render (RNF-008).
package server

import (
	"context"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/mpraes/archi/internal/ai"
	"github.com/mpraes/archi/internal/model"
	"github.com/mpraes/archi/web"
)

type Server struct {
	summary model.Summary
	ai      *ai.Config
	logger  *slog.Logger
}

// New creates a server for the given metrics summary. aiCfg may be nil when AI
// is disabled (offline default).
func New(s model.Summary, aiCfg *ai.Config, logger *slog.Logger) *Server {
	return &Server{summary: s, ai: aiCfg, logger: logger}
}

// Run serves HTTP on ln until ctx is cancelled.
func (s *Server) Run(ctx context.Context, ln net.Listener) error {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer, requestLogger(s.logger))

	r.Get("/api/metrics", s.handleMetrics)
	r.Get("/api/modules/{module}", s.handleModule)
	r.Get("/api/ai/insights", s.handleAIInsights)
	r.Get("/api/ai/enabled", s.handleAIEnabled)

	dist, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		return err
	}
	r.Handle("/*", http.FileServer(http.FS(dist)))

	srv := &http.Server{Handler: r}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.summary)
}

func (s *Server) handleModule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "module")
	for _, m := range s.summary.Modules {
		if m.Module == name {
			writeJSON(w, m)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) handleAIEnabled(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]bool{"enabled": s.ai != nil})
}

func (s *Server) handleAIInsights(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		http.Error(w, "IA desativada (use --ai ou defina GEMINI_API_KEY)", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming não suportado", http.StatusInternalServerError)
		return
	}
	stream, err := ai.Insights(r.Context(), s.ai, s.summary)
	if err != nil {
		writeSSE(w, "error", err.Error())
		return
	}
	for chunk := range stream {
		writeSSE(w, "chunk", chunk)
		flusher.Flush()
	}
	writeSSE(w, "done", "")
	flusher.Flush()
}

func writeSSE(w http.ResponseWriter, event, data string) {
	var b strings.Builder
	b.WriteString("event: ")
	b.WriteString(event)
	b.WriteString("\ndata: ")
	_ = json.NewEncoder(&b).Encode(data)
	b.WriteByte('\n')
	_, _ = w.Write([]byte(b.String()))
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debug("http", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}
