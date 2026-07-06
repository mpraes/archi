package parser

import (
	"path/filepath"
	"testing"
)

func TestParseJSImportsDetails(t *testing.T) {
	src := []byte(`
import foo from "./helper.js";
import { bar as baz } from "../shared/util";
const x = require("./local");
const dyn = await import("./dynamic");
`)
	imports, aliases := parseJSImportsDetails(src)
	if len(imports) < 3 {
		t.Fatalf("imports=%v", imports)
	}
	if aliases["x"] != "./local" {
		t.Fatalf("aliases=%v", aliases)
	}
}

func TestCallRootToken(t *testing.T) {
	cases := map[string]string{
		"await foo.bar()": "foo",
		"new Thing()":     "Thing",
		"plain":           "plain",
	}
	for in, want := range cases {
		if got := callRootToken(in); got != want {
			t.Fatalf("%q => %q want %q", in, got, want)
		}
	}
}

func TestNormalizeJSImport(t *testing.T) {
	p := &jsTSParser{root: "/proj"}
	from := filepath.Join("/proj", "src", "index.js")
	if got := p.normalizeJSImport("./helper.js", from); got != "src" {
		t.Fatalf("relative = %q", got)
	}
	if got := p.normalizeJSImport("@/utils/foo", from); got != "utils/foo" {
		t.Fatalf("alias = %q", got)
	}
	if got := p.normalizeJSImport("/src/pkg", from); got != "src" {
		t.Fatalf("absolute = %q", got)
	}
}

func TestJsModuleLanguageAndCountLines(t *testing.T) {
	if jsModuleLanguage("a.ts") != "ts" || jsModuleLanguage("a.js") != "js" {
		t.Fatal("js module language")
	}
	if countLines([]byte("a\nb\nc")) != 3 {
		t.Fatal("countLines")
	}
}

func TestJsTSParserModuleFor(t *testing.T) {
	p := &jsTSParser{root: "/proj"}
	path := filepath.Join("/proj", "src", "index.js")
	if p.moduleFor(path) != "src" {
		t.Fatalf("moduleFor=%q", p.moduleFor(path))
	}
}

func TestJsResolveCallTarget(t *testing.T) {
	p := &jsTSParser{root: "/proj"}
	mods := map[string]struct{}{"src": {}, "lib": {}}
	p.aliasesByFile = map[string]map[string]string{
		"x": {"helper": "./lib"},
	}
	if got := p.resolveCallTarget("lib.run", "src", mods); got != "lib" {
		t.Fatalf("got %q", got)
	}
}
