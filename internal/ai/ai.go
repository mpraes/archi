// Package ai provides optional Gemini-based architectural insights.
//
// It is opt-in and never blocks initial UI render: the frontend opens with
// the local metrics immediately and streams AI text in via the server's
// /api/ai/insights SSE endpoint (RNF-008).
//
// The metrics payload sent to the model contains NO source code, only
// numeric metrics, graph topology and alerts (RFD-011, RNF-009). API keys
// are taken only from --api-key or GEMINI_API_KEY and are never persisted.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"

	"github.com/mpraes/archi/internal/model"

	"google.golang.org/genai"
)

// Config carries the API key. A nil Config means AI is disabled.
type Config struct {
	APIKey string
	Model  string // defaults to gemini-2.5-flash
	Logger *slog.Logger
}

// Insights streams textual architectural recommendations for the top critical
// hotspots. The returned iterator yields text chunks; it is closed when the
// stream ends or ctx is cancelled.
func Insights(ctx context.Context, cfg *Config, s model.Summary) (iter.Seq2[string, error], error) {
	if cfg == nil || cfg.APIKey == "" {
		return nil, fmt.Errorf("IA desativada: chave de API ausente")
	}
	modelName := cfg.Model
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("cliente Gemini: %w", err)
	}

	prompt := buildPrompt(s)
	contents := []*genai.Content{{Parts: []*genai.Part{{Text: prompt}}}}

	out := make(chan string, 32)
	errOut := make(chan error, 1)
	go func() {
		defer close(out)
		stream := client.Models.GenerateContentStream(ctx, modelName, contents, &genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: systemPrompt()}}},
		})
		for resp, err := range stream {
			if err != nil {
				errOut <- err
				return
			}
			if resp == nil {
				continue
			}
			if text := resp.Text(); text != "" {
				select {
				case out <- text:
				case <-ctx.Done():
					return
				}
			}
		}
		errOut <- nil
	}()

	return func(yield func(string, error) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-out:
				if !ok {
					if e := <-errOut; e != nil {
						yield("", e)
					}
					return
				}
				if !yield(chunk, nil) {
					return
				}
			}
		}
	}, nil
}

func systemPrompt() string {
	return "Você é um consultor sênior de arquitetura de software. Analise métricas " +
		"estruturais e forneça recomendações acionáveis, em português brasileiro, " +
		"humanas e diretas. NUNCA invente nomes de arquivos que não estejam nos " +
		"dados. Foque nos 3 pontos mais críticos. Não inclua código-fonte."
}

func buildPrompt(s model.Summary) string {
	// Serialize only metrics topology — no source code (RNF-009).
	payload := struct {
		Project  string                  `json:"project"`
		Modules  []model.ModuleMetrics   `json:"modules"`
		Hotspots []string                `json:"hotspots"`
		Connasc  []model.Connascence     `json:"connascence"`
	}{
		Project:  s.ProjectName,
		Modules:  s.Modules,
		Hotspots: s.Hotspots,
		Connasc:  s.Connascence,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "Métricas indisponíveis."
	}
	return "A seguir estão métricas arquiteturais (sem código-fonte) de um projeto. " +
		"Identifique os 3 pontos mais críticos e recomende ações concretas:\n\n" + string(b)
}