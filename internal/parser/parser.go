// Package parser discovers source files, resolves module boundaries from
// manifests (go.mod, package.json), and dispatches to a language parser.
//
// For Go targets the stdlib go/ast based parser is used (preferred over
// tree-sitter per docs/stack.md). Tree-sitter integration is the extension
// point for other languages.
package parser

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mpraes/archi/internal/model"
)

// readGoModulePath extracts the module path from a go.mod at root, if present.
func readGoModulePath(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
}

// Options controls a parse run.
type Options struct {
	Lang    string   // force language; "" auto-detect
	Exclude []string // extra glob/dir patterns to skip
	ModulePath string // resolved module path (for Go)
}

// Result of a parse run plus any per-file warnings (never fatal).
type Result struct {
	Program *model.Program
	Warnings []string
}

// Parse walks root and parses all supported source files. It never fails on a
// single broken file: warnings are collected and the file is skipped (RNF-003).
func Parse(root string, opts Options) *Result {
	root = abs(root)
	lang := opts.Lang
	if lang == "" {
		lang = detectLang(root)
	}
	modulePath := opts.ModulePath
	if modulePath == "" {
		modulePath = readGoModulePath(root)
	}

	res := &Result{Program: &model.Program{ProjectName: filepath.Base(root), ModulePath: modulePath}}
	var files []string
	skip := buildSkip(opts)

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			res.Warnings = append(res.Warnings, "walk "+path+": "+err.Error())
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if skipDir(base, path, root, skip) {
				return filepath.SkipDir
			}
			return nil
		}
		if skipFile(path, skip) {
			return nil
		}
		if langMatch(path, lang) {
			files = append(files, path)
		}
		return nil
	})

	// Dispatch by language.
	var p langParser
	switch lang {
	case "go", "":
		p = &goParser{root: root, modulePath: modulePath, warnings: &res.Warnings}
	case "ts", "js":
		// Tree-sitter based parser is the extension point for non-Go targets.
		res.Warnings = append(res.Warnings, "language "+lang+" requires tree-sitter grammar support (not yet wired); skipping")
		return res
	default:
		res.Warnings = append(res.Warnings, "unsupported language: "+lang)
		return res
	}

	for _, f := range files {
		pf, err := p.parseFile(f)
		if err != nil {
			res.Warnings = append(res.Warnings, "parse "+f+": "+err.Error())
			continue
		}
		res.Program.Files = append(res.Program.Files, pf)
	}
	res.Program.Modules = p.modules()
	// Resolve cross-module calls now that all modules are known.
	p.resolve(res.Program)
	return res
}

type langParser interface {
	parseFile(path string) (model.File, error)
	modules() []model.Module
	resolve(*model.Program)
}

func detectLang(root string) string {
	if fileExists(filepath.Join(root, "go.mod")) {
		return "go"
	}
	if fileExists(filepath.Join(root, "package.json")) {
		return "ts"
	}
	// Fallback: sniff first supported file.
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go":
			found = "go"
			return filepath.SkipAll
		case ".ts", ".tsx":
			found = "ts"
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func abs(p string) string {
	a, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return a
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// buildSkip merges default excludes with user extras.
func buildSkip(opts Options) []string {
	def := []string{
		"node_modules", ".git", "vendor", "dist", "build", "out",
		"*_test.go", "*.spec.ts", "*.spec.js", "*.spec.tsx",
	}
	return append(append([]string{}, def...), opts.Exclude...)
}

// skipDir reports a directory to skip entirely (by base name or matched pattern).
func skipDir(base, path, root string, skip []string) bool {
	for _, pat := range skip {
		// Directory patterns match by base name (no glob slash).
		if !strings.ContainsAny(pat, "*?[") {
			if base == pat {
				return true
			}
			continue
		}
		// Glob patterns apply to the full path relative-to-root.
		rel, err := filepath.Rel(root, path)
		if err == nil && matchGlob(rel, pat) {
			return true
		}
		// Also try matching just the base name against glob patterns.
		if matchGlob(base, pat) {
			return true
		}
	}
	// Hidden directories.
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}
	return false
}

func skipFile(path string, skip []string) bool {
	base := filepath.Base(path)
	for _, pat := range skip {
		if matchGlob(base, pat) {
			return true
		}
	}
	return false
}

func langMatch(path, lang string) bool {
	switch lang {
	case "go":
		return filepath.Ext(path) == ".go"
	case "ts", "js":
		e := filepath.Ext(path)
		return e == ".ts" || e == ".tsx" || e == ".js" || e == ".jsx"
	}
	return false
}

func matchGlob(name, pattern string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}