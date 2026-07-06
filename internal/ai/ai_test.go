package ai

import (
	"strings"
	"testing"

	"github.com/mpraes/archi/internal/model"
)

func TestBuildPromptOmitsSource(t *testing.T) {
	s := model.Summary{
		ProjectName: "demo",
		Modules: []model.ModuleMetrics{{
			Module: "alpha", Distance: 0.5, Instability: 0.5,
		}},
		Hotspots: []string{"alpha"},
	}
	prompt := buildPrompt(s)
	if !strings.Contains(prompt, "demo") {
		t.Fatal("missing project name")
	}
	for _, forbidden := range []string{"func ", "package ", ".go", "import "} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt may contain source-like content: %q", forbidden)
		}
	}
}

func TestSystemPrompt(t *testing.T) {
	if systemPrompt() == "" {
		t.Fatal("empty system prompt")
	}
}

func TestInsightsRequiresAPIKey(t *testing.T) {
	_, err := Insights(t.Context(), nil, model.Summary{})
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	_, err = Insights(t.Context(), &Config{}, model.Summary{})
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}
