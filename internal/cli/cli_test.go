package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mpraes/archi/internal/model"
)

func TestPathArg(t *testing.T) {
	g := &GlobalFlags{}
	if g.pathArg(nil) != "." {
		t.Fatal("default path")
	}
	if g.pathArg([]string{"foo"}) != "foo" {
		t.Fatal("explicit path")
	}
}

func TestExportJSON(t *testing.T) {
	var buf bytes.Buffer
	s := model.Summary{ProjectName: "p", ModuleCount: 0, Modules: []model.ModuleMetrics{}, Hotspots: []string{}, Connascence: []model.Connascence{}}
	if err := exportJSON(&buf, s); err != nil {
		t.Fatal(err)
	}
	var decoded model.Summary
	if err := json.NewDecoder(&buf).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ProjectName != "p" {
		t.Fatalf("project = %q", decoded.ProjectName)
	}
}

func TestExportMarkdown(t *testing.T) {
	var buf bytes.Buffer
	s := model.Summary{
		ProjectName: "demo",
		ModuleCount: 1,
		Modules: []model.ModuleMetrics{{
			Module: "m", Afferent: 1, Efferent: 2,
			Instability: 0.5, Abstraction: 0.5, Distance: 0.1,
		}},
		Hotspots:    []string{"m"},
		Connascence: []model.Connascence{{Kind: "name", From: "a", To: "b", Detail: "x"}},
	}
	if err := exportMarkdown(&buf, s); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"# Archi", "## Hotspots", "## Módulos", "## Conascência"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestScanCoreMiniGo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	g := &GlobalFlags{Lang: "go"}
	summary, warnings := scanCore(root, g)
	if summary.ModuleCount < 2 {
		t.Fatalf("modules = %d", summary.ModuleCount)
	}
	if len(warnings) == 0 {
		t.Log("no warnings from mini-go fixture")
	}
}

func TestCheckViolations(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	cmd := NewRootCmd("test")
	cmd.SetArgs([]string{"check", "--max-distance", "0.5", root})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected violation error")
	}
	var ec *exitCodeError
	if !errors.As(err, &ec) || ec.code != 1 {
		t.Fatalf("err = %#v", err)
	}
}

func TestCheckOK(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	cmd := NewRootCmd("test")
	cmd.SetArgs([]string{"check", "--max-distance", "1.0", root})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestExportCmdJSON(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	cmd := NewRootCmd("test")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"export", root, "--format", "json", "--lang", "go"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestExportInvalidFormat(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	cmd := NewRootCmd("test")
	cmd.SetArgs([]string{"export", "--format", "xml", "--lang", "go", root})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected format error")
	}
}
