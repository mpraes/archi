package analyzer

import (
	"testing"

	"github.com/mpraes/archi/internal/model"
)

func TestAnalyzeCouplingAndMetrics(t *testing.T) {
	p := &model.Program{
		ProjectName: "test",
		Modules: []model.Module{
			{Name: "alpha", Path: "alpha", Language: "go", Files: 1, Imports: map[string]struct{}{"beta": {}}},
			{Name: "beta", Path: "beta", Language: "go", Files: 1},
		},
		Files: []model.File{
			{
				Path: "alpha/a.go", Module: "alpha",
				Blocks: []model.Block{
					{Name: "Alpha", Kind: model.BlockFunc, Complexity: 1, Calls: []model.Call{{Target: "beta.Beta"}}},
				},
			},
			{
				Path: "beta/b.go", Module: "beta",
				Blocks: []model.Block{
					{Name: "Beta", Kind: model.BlockFunc, Complexity: 1},
				},
			},
		},
	}

	s := Analyze(p)
	if s.ModuleCount != 2 {
		t.Fatalf("module count = %d, want 2", s.ModuleCount)
	}
	if p.Modules[0].Efferent != 1 || p.Modules[1].Afferent != 1 {
		t.Fatalf("coupling mismatch: alpha eff=%d beta aff=%d", p.Modules[0].Efferent, p.Modules[1].Afferent)
	}
	if p.Modules[0].Instability <= 0 || p.Modules[1].Instability >= 1 {
		t.Fatalf("instability out of range: alpha=%f beta=%f", p.Modules[0].Instability, p.Modules[1].Instability)
	}
}

func TestAnalyzeGodBlocksAndOrphans(t *testing.T) {
	p := &model.Program{
		Modules: []model.Module{{Name: "m", Path: "m", Language: "go", Files: 1}},
		Files: []model.File{{
			Path: "m/m.go", Module: "m",
			Blocks: []model.Block{
				{Name: "God", Kind: model.BlockFunc, Complexity: 20},
				{Name: "Orphan", Kind: model.BlockFunc, Complexity: 1},
				{Name: "Main", Kind: model.BlockFunc, Complexity: 1, Calls: []model.Call{{Target: "God"}}},
			},
		}},
	}
	Analyze(p)
	m := p.Modules[0]
	if len(m.GodBlocks) != 1 || m.GodBlocks[0] != "God" {
		t.Fatalf("god blocks = %#v", m.GodBlocks)
	}
	if len(m.OrphanBlocks) != 2 {
		t.Fatalf("orphan blocks = %#v", m.OrphanBlocks)
	}
}

func TestAnalyzeConnascenceMeaning(t *testing.T) {
	p := &model.Program{
		Modules: []model.Module{
			{Name: "a", Path: "a", Language: "go", Files: 1},
			{Name: "b", Path: "b", Language: "go", Files: 1},
		},
		Files: []model.File{
			{Path: "a/a.go", Module: "a", Literals: []model.Literal{{Kind: model.LitString, Value: "shared-literal", Line: 1}}},
			{Path: "b/b.go", Module: "b", Literals: []model.Literal{{Kind: model.LitString, Value: "shared-literal", Line: 2}}},
		},
	}
	s := Analyze(p)
	found := false
	for _, c := range s.Connascence {
		if c.Kind == "meaning" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected meaning connascence")
	}
}

func TestAnalyzeHotspotsSkipsBoilerplate(t *testing.T) {
	p := &model.Program{
		Modules: []model.Module{
			{Name: "template", Path: "template", Language: "go", Instability: 0.9, Abstraction: 0.1, Distance: 0.8},
			{Name: "core", Path: "core", Language: "go", Instability: 0.9, Abstraction: 0.1, Distance: 0.8},
		},
	}
	// Pre-set distance fields are overwritten; use files with god blocks instead.
	p.Files = []model.File{{
		Path: "core/c.go", Module: "core",
		Blocks: []model.Block{{Name: "Huge", Kind: model.BlockFunc, Complexity: 20}},
	}}
	s := Analyze(p)
	for _, h := range s.Hotspots {
		if h == "template" {
			t.Fatal("boilerplate module should not be hotspot")
		}
	}
}

func TestRoundAndQuote(t *testing.T) {
	if got := round(0.123456); got != 0.123 {
		t.Fatalf("round = %f", got)
	}
	long := quote("abcdefghijklmnopqrstuvwxyz0123456789extra")
	if len(long) > 45 {
		t.Fatalf("quote too long: %q", long)
	}
}

func TestIsBoilerplateModule(t *testing.T) {
	if !isBoilerplateModule(model.Module{Name: "foo-template", Path: "x"}) {
		t.Fatal("expected boilerplate")
	}
	if isBoilerplateModule(model.Module{Name: "core", Path: "core"}) {
		t.Fatal("unexpected boilerplate")
	}
}

func TestBuildSummaryNilSlices(t *testing.T) {
	p := &model.Program{
		Modules: []model.Module{{Name: "m", Path: "m", Language: "go"}},
	}
	s := buildSummary(p, nil, nil)
	if s.Hotspots == nil || s.Connascence == nil {
		t.Fatal("expected empty slices, not nil")
	}
	if len(s.Modules) != 1 {
		t.Fatalf("modules = %d", len(s.Modules))
	}
	if s.Modules[0].OrphanBlocks == nil || s.Modules[0].GodBlocks == nil {
		t.Fatal("module slices should be non-nil")
	}
}

func TestConnascenceNameCrossModule(t *testing.T) {
	p := &model.Program{
		Modules: []model.Module{
			{Name: "client", Path: "client", Language: "go"},
			{Name: "server", Path: "server", Language: "go"},
		},
		Files: []model.File{
			{
				Path: "client/c.go", Module: "client",
				Blocks: []model.Block{
					{Name: "Run", Kind: model.BlockFunc, Calls: []model.Call{{Target: "Svc.Do"}}},
				},
			},
			{
				Path: "server/s.go", Module: "server",
				Blocks: []model.Block{
					{Name: "Svc", Kind: model.BlockType, Receiver: "struct"},
					{Name: "Do", Kind: model.BlockMethod, Receiver: "Svc"},
				},
			},
		},
	}
	byName := map[string]*model.Module{}
	for i := range p.Modules {
		byName[p.Modules[i].Name] = &p.Modules[i]
	}
	out := connascenceName(p, byName)
	if len(out) == 0 {
		t.Fatal("expected name connascence")
	}
}
