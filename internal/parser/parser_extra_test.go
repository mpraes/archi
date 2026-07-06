package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLangFixtures(t *testing.T) {
	root, err := repoTestdata()
	if err != nil {
		t.Fatal(err)
	}
	if got := detectLang(filepath.Join(root, "mini-js")); got != "js" {
		t.Fatalf("mini-js lang = %q", got)
	}
	if got := detectLang(filepath.Join(root, "mini-py")); got != "py" {
		t.Fatalf("mini-py lang = %q", got)
	}
	if got := detectLang(root); got != "all" {
		t.Fatalf("testdata lang = %q", got)
	}
}

func TestSkipDirGlobPattern(t *testing.T) {
	root := t.TempDir()
	custom := filepath.Join(root, "custom_skip")
	if err := os.MkdirAll(custom, 0o755); err != nil {
		t.Fatal(err)
	}
	skip := buildSkip(Options{Exclude: []string{"custom_skip"}})
	if !skipDir("custom_skip", custom, root, skip) {
		t.Fatal("expected custom skip dir")
	}
}

func TestGoParserResolveModules(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	res := Parse(root, Options{Lang: "go", ModulePath: "github.com/example/mini-go"})
	if len(res.Program.Modules) < 2 {
		t.Fatalf("modules=%d", len(res.Program.Modules))
	}
	for _, f := range res.Program.Files {
		if f.Module == "" {
			t.Fatalf("empty module for %s", f.Path)
		}
	}
}

func TestParseInvalidGoFileInTemp(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.go")
	if err := os.WriteFile(bad, []byte("package x\n<<<invalid syntax>>>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := Parse(dir, Options{Lang: "go"})
	if len(res.Warnings) == 0 && len(res.Program.Files) > 0 {
		t.Log("parser recovered invalid file without warning")
	}
	if len(res.Program.Files) == 0 && len(res.Warnings) == 0 {
		t.Fatal("expected warning or parsed recovery")
	}
}

func repoTestdata() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "..", "..", "testdata"), nil
}
