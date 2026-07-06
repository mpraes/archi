package parser

import "testing"

func TestParsePythonImportsDetails(t *testing.T) {
	src := []byte(`
import os, json as js
from .utils import helper
from pkg.sub import Thing
`)
	imports, aliases := parsePythonImportsDetails(src)
	if len(imports) == 0 {
		t.Fatalf("imports=%v aliases=%v", imports, aliases)
	}
}

func TestParsePyImportItem(t *testing.T) {
	mod, tok := parsePyImportItem("json as js")
	if mod != "json" || tok != "js" {
		t.Fatalf("got mod=%q tok=%q", mod, tok)
	}
}

func TestResolveRelativePyImport(t *testing.T) {
	if got := resolveRelativePyImport(".helper", "pkg/mod"); got != "pkg/mod/helper" {
		t.Fatalf("got %q", got)
	}
	if got := resolveRelativePyImport("..sibling", "pkg/sub"); got != "pkg/sibling" {
		t.Fatalf("parent got %q", got)
	}
}

func TestNormalizeKnownModule(t *testing.T) {
	mods := map[string]struct{}{"pkg/sub": {}, "utils": {}}
	if got := normalizeKnownModule("pkg/sub/mod", mods); got != "pkg/sub" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeKnownModule("utils", mods); got != "utils" {
		t.Fatalf("got %q", got)
	}
}

func TestPyParserModuleFor(t *testing.T) {
	p := &pyParser{root: "/proj"}
	if p.moduleFor("/proj/main.py") != "root" {
		t.Fatal("root module")
	}
}

func TestResolveCallByToken(t *testing.T) {
	mods := map[string]struct{}{"utils": {}}
	if got := resolveCallByToken("utils.helper", "app", mods); got != "utils" {
		t.Fatalf("got %q", got)
	}
}
