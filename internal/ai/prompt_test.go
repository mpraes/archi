package ai

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mpraes/archi/internal/model"
)

func TestBuildPromptIncludesMetricsJSON(t *testing.T) {
	s := model.Summary{
		ProjectName: "proj",
		Modules: []model.ModuleMetrics{
			{Module: "a", Distance: 0.42, Instability: 0.5, Abstraction: 0.5},
		},
		Hotspots:    []string{"a"},
		Connascence: []model.Connascence{{Kind: "name", From: "a", To: "b", Detail: "calls"}},
	}
	prompt := buildPrompt(s)
	var payload struct {
		Project string `json:"project"`
	}
	start := strings.Index(prompt, "{")
	if start < 0 {
		t.Fatal("missing json payload")
	}
	if err := json.Unmarshal([]byte(prompt[start:]), &payload); err != nil {
		t.Fatalf("payload json: %v", err)
	}
	if payload.Project != "proj" {
		t.Fatalf("project = %q", payload.Project)
	}
}

func TestBuildPromptMarshalFallback(t *testing.T) {
	// Ensure buildPrompt never returns empty string.
	prompt := buildPrompt(model.Summary{ProjectName: "x"})
	if prompt == "" {
		t.Fatal("empty prompt")
	}
}

func TestConfigModelDefault(t *testing.T) {
	cfg := &Config{APIKey: "k"}
	if cfg.Model != "" {
		t.Fatal("default model should be empty until Insights")
	}
}
