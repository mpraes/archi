package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoParserIntegration(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "mini-go")
	res := Parse(root, Options{Lang: "go", ModulePath: "github.com/example/mini-go"})
	if len(res.Program.Files) < 2 {
		t.Fatalf("files = %d", len(res.Program.Files))
	}
	names := map[string]bool{}
	for _, m := range res.Program.Modules {
		names[m.Name] = true
	}
	if !names["github.com/example/mini-go/alpha"] && !names["alpha"] {
		t.Fatalf("modules = %#v", res.Program.Modules)
	}
}

func TestJSParserIntegration(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("tree-sitter requires CGO")
	}
	root := filepath.Join("..", "..", "testdata", "mini-js")
	res := Parse(root, Options{Lang: "js"})
	if len(res.Program.Files) == 0 {
		t.Fatal("expected parsed JS files")
	}
}

func TestPythonParserIntegration(t *testing.T) {
	if os.Getenv("CGO_ENABLED") == "0" {
		t.Skip("tree-sitter requires CGO")
	}
	root := filepath.Join("..", "..", "testdata", "mini-py")
	res := Parse(root, Options{Lang: "py"})
	if len(res.Program.Files) == 0 {
		t.Fatal("expected parsed Python files")
	}
}
