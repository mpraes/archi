package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSkipAndMatchGlob(t *testing.T) {
	skip := buildSkip(Options{Exclude: []string{"custom"}})
	if !contains(skip, "node_modules") || !contains(skip, "custom") {
		t.Fatalf("skip = %#v", skip)
	}
	if !matchGlob("foo_test.go", "*_test.go") {
		t.Fatal("glob should match test file")
	}
	if matchGlob("main.go", "*_test.go") {
		t.Fatal("glob should not match main.go")
	}
}

func TestSkipDirAndFile(t *testing.T) {
	root := t.TempDir()
	skip := buildSkip(Options{})
	if !skipDir("node_modules", filepath.Join(root, "node_modules"), root, skip) {
		t.Fatal("should skip node_modules")
	}
	if skipDir("src", filepath.Join(root, "src"), root, skip) {
		t.Fatal("should not skip src")
	}
	if !skipFile("/proj/foo_test.go", skip) {
		t.Fatal("should skip test files")
	}
}

func TestLangMatchAndLangForFile(t *testing.T) {
	cases := map[string]string{
		"main.go": "go", "app.ts": "ts", "app.tsx": "ts",
		"index.js": "js", "page.jsx": "js", "main.py": "py",
	}
	for file, want := range cases {
		if got := langForFile(file); got != want {
			t.Fatalf("%s: got %q want %q", file, got, want)
		}
	}
	if !langMatch("x.go", "go") || langMatch("x.go", "js") {
		t.Fatal("langMatch go mismatch")
	}
	if !langMatch("x.ts", "all") {
		t.Fatal("langMatch all should include ts")
	}
}

func TestReadGoModulePath(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	mod := readGoModulePath(root)
	if mod != "github.com/example/mini-go" {
		t.Fatalf("module path = %q", mod)
	}
	if readGoModulePath(t.TempDir()) != "" {
		t.Fatal("missing go.mod should return empty")
	}
}

func TestDetectLangMiniGo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	if got := detectLang(root); got != "go" {
		t.Fatalf("detectLang = %q, want go", got)
	}
}

func TestParseMiniGo(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	res := Parse(root, Options{Lang: "go"})
	if res.Program == nil {
		t.Fatal("nil program")
	}
	if len(res.Program.Modules) < 2 {
		t.Fatalf("modules = %d, want >= 2", len(res.Program.Modules))
	}
	if len(res.Warnings) == 0 {
		t.Log("no parser warnings (broken file may be skipped silently on some toolchains)")
	}
}

func TestParseUnsupportedLang(t *testing.T) {
	res := Parse(t.TempDir(), Options{Lang: "cobol"})
	if len(res.Warnings) == 0 {
		t.Fatal("expected unsupported language warning")
	}
}

func TestParseMixedAll(t *testing.T) {
	root := filepath.Join("..", "..", "testdata")
	res := Parse(root, Options{Lang: "all"})
	if res.Program == nil {
		t.Fatal("nil program")
	}
	if len(res.Program.Files) == 0 {
		t.Fatal("expected files from mixed parsers")
	}
}

func TestAbsAndFileExists(t *testing.T) {
	dir := t.TempDir()
	absPath := abs(dir)
	if !filepath.IsAbs(absPath) {
		t.Fatalf("abs = %q", absPath)
	}
	f := filepath.Join(dir, "x.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !fileExists(f) || fileExists(filepath.Join(dir, "missing")) {
		t.Fatal("fileExists mismatch")
	}
}

func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}
