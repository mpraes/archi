package parser

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"

	"github.com/mpraes/archi/internal/model"
)

var (
	importFromRe      = regexp.MustCompile(`(?m)^\s*import(?:[\s\w*\{\},$]*from\s*)?["']([^"']+)["']`)
	importClauseRe    = regexp.MustCompile(`(?m)^\s*import\s+(.+?)\s+from\s+["']([^"']+)["']`)
	requireRe         = regexp.MustCompile(`(?m)\brequire\(\s*["']([^"']+)["']\s*\)`)
	requireAssignRe   = regexp.MustCompile(`(?m)\b(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*require\(\s*["']([^"']+)["']\s*\)`)
	dynImportRe       = regexp.MustCompile(`(?m)\bimport\(\s*["']([^"']+)["']\s*\)`)
	namedImportItemRe = regexp.MustCompile(`\s+as\s+`)
)

type jsTSParser struct {
	root     string
	warnings *[]string

	files         []model.File
	modulesList   []model.Module
	aliasesByFile map[string]map[string]string
}

func (p *jsTSParser) parseFile(path string) (model.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return model.File{}, err
	}

	lang, err := languageFor(path)
	if err != nil {
		return model.File{}, err
	}

	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(lang); err != nil {
		return model.File{}, err
	}

	tree := parser.Parse(src, nil)
	defer tree.Close()
	root := tree.RootNode()

	imports, aliases := parseJSImportsDetails(src)
	if p.aliasesByFile == nil {
		p.aliasesByFile = map[string]map[string]string{}
	}
	p.aliasesByFile[path] = aliases

	f := model.File{
		Path:     path,
		Module:   p.moduleFor(path),
		Lines:    countLines(src),
		Imports:  imports,
		Blocks:   collectJSBlocks(root, src),
		Literals: collectJSLiterals(root, src),
	}
	if root.HasError() {
		f.Errors = append(f.Errors, "syntax errors detected; partial tree-sitter parse used")
	}

	p.files = append(p.files, f)
	return f, nil
}

func (p *jsTSParser) modules() []model.Module {
	indexByName := map[string]int{}
	rawImportsByModule := map[string][]string{}
	for _, f := range p.files {
		idx, ok := indexByName[f.Module]
		if !ok {
			mpath := p.root
			if f.Module != "root" {
				mpath = filepath.Join(p.root, filepath.FromSlash(f.Module))
			}
			p.modulesList = append(p.modulesList, model.Module{
				Name:     f.Module,
				Path:     mpath,
				Language: jsModuleLanguage(f.Path),
				Imports:  map[string]struct{}{},
			})
			idx = len(p.modulesList) - 1
			indexByName[f.Module] = idx
		}
		m := &p.modulesList[idx]
		fileLang := jsModuleLanguage(f.Path)
		if m.Language == "" {
			m.Language = fileLang
		} else if m.Language != fileLang && m.Language != "js+ts" {
			m.Language = "js+ts"
		}
		m.Files++
		rawImportsByModule[f.Module] = append(rawImportsByModule[f.Module], f.Imports...)
		for _, b := range f.Blocks {
			if b.Kind == model.BlockType {
				if b.Receiver == "interface" {
					m.Abstracts++
				} else {
					m.Concretes++
				}
			}
			m.SumCyclo += b.Complexity
			if b.Complexity > m.MaxCyclo {
				m.MaxCyclo = b.Complexity
			}
		}
	}
	known := map[string]struct{}{}
	for _, m := range p.modulesList {
		known[m.Name] = struct{}{}
	}
	for i := range p.modulesList {
		m := &p.modulesList[i]
		for _, im := range rawImportsByModule[m.Name] {
			target := p.normalizeJSImport(im, m.Path)
			if target == "" || target == m.Name {
				continue
			}
			if resolved := normalizeKnownModule(target, known); resolved != "" && resolved != m.Name {
				m.Imports[resolved] = struct{}{}
			}
		}
	}
	return p.modulesList
}

func (p *jsTSParser) resolve(prog *model.Program) {
	byName := map[string]struct{}{}
	for _, m := range prog.Modules {
		byName[m.Name] = struct{}{}
	}
	parsed := map[string]struct{}{}
	for _, f := range p.files {
		parsed[f.Path] = struct{}{}
	}
	for i := range prog.Files {
		f := &prog.Files[i]
		if _, ok := parsed[f.Path]; !ok {
			continue
		}
		aliases := p.aliasesByFile[f.Path]
		for bi := range f.Blocks {
			b := &f.Blocks[bi]
			for ci := range b.Calls {
				root := callRootToken(b.Calls[ci].Target)
				if rawImport, ok := aliases[root]; ok {
					if resolved := normalizeKnownModule(p.normalizeJSImport(rawImport, f.Path), byName); resolved != "" {
						b.Calls[ci].ResolvedModule = resolved
						continue
					}
				}
				target := p.resolveCallTarget(b.Calls[ci].Target, f.Module, byName)
				b.Calls[ci].ResolvedModule = target
			}
		}
	}
}

func (p *jsTSParser) resolveCallTarget(target, from string, modules map[string]struct{}) string {
	token := callRootToken(target)
	if token == "" || token == "this" {
		return ""
	}
	for name := range modules {
		if name == from {
			continue
		}
		if lastElem(name) == token {
			return name
		}
	}
	return ""
}

func callRootToken(target string) string {
	target = strings.TrimSpace(target)
	target = strings.TrimPrefix(target, "await ")
	if i := strings.Index(target, "("); i >= 0 {
		target = target[:i]
	}
	target = strings.TrimPrefix(target, "new ")
	if i := strings.Index(target, "."); i >= 0 {
		return target[:i]
	}
	return target
}

func (p *jsTSParser) moduleFor(filePath string) string {
	rel, err := filepath.Rel(p.root, filePath)
	if err != nil {
		return "root"
	}
	dir := filepath.ToSlash(filepath.Dir(rel))
	if dir == "." || dir == "" {
		return "root"
	}
	return dir
}

func (p *jsTSParser) normalizeJSImport(raw, fromPath string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "/") {
		raw = strings.TrimPrefix(raw, "/")
		if raw == "" {
			return "root"
		}
		return strings.Trim(filepath.ToSlash(path.Dir(raw)), "/")
	}
	if !strings.HasPrefix(raw, ".") {
		raw = strings.TrimPrefix(raw, "@/")
		raw = strings.TrimPrefix(raw, "~/")
		raw = strings.Trim(raw, "/")
		raw = strings.ReplaceAll(raw, ".", "/")
		return raw
	}
	base := fromPath
	if filepath.Ext(base) != "" {
		base = filepath.Dir(base)
	}
	targetAbs := filepath.Clean(filepath.Join(base, raw))
	rel, err := filepath.Rel(p.root, targetAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ""
	}

	relSlash := filepath.ToSlash(rel)
	ext := filepath.Ext(relSlash)
	switch {
	case ext != "":
		relSlash = path.Dir(relSlash)
	case strings.HasSuffix(relSlash, "/index"):
		relSlash = path.Dir(relSlash)
	case relSlash == "index":
		relSlash = "."
	}
	if relSlash == "." || relSlash == "" {
		return "root"
	}
	return strings.TrimPrefix(relSlash, "./")
}

func languageFor(filePath string) (*sitter.Language, error) {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".js", ".jsx":
		return sitter.NewLanguage(tree_sitter_javascript.Language()), nil
	case ".tsx":
		return sitter.NewLanguage(tree_sitter_typescript.LanguageTSX()), nil
	default:
		return sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()), nil
	}
}

func jsModuleLanguage(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".ts", ".tsx":
		return "ts"
	default:
		return "js"
	}
}

func parseJSImportsDetails(src []byte) ([]string, map[string]string) {
	text := string(src)
	var out []string
	seen := map[string]struct{}{}
	aliases := map[string]string{}
	addImport := func(im string) {
		im = strings.TrimSpace(im)
		if im == "" {
			return
		}
		if _, ok := seen[im]; ok {
			return
		}
		seen[im] = struct{}{}
		out = append(out, im)
	}

	appendMatches := func(re *regexp.Regexp) {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			if len(m) < 2 {
				continue
			}
			addImport(m[1])
		}
	}
	appendMatches(importFromRe)
	appendMatches(requireRe)
	appendMatches(dynImportRe)

	for _, m := range requireAssignRe.FindAllStringSubmatch(text, -1) {
		if len(m) < 3 {
			continue
		}
		token := strings.TrimSpace(m[1])
		im := strings.TrimSpace(m[2])
		if token != "" && im != "" {
			aliases[token] = im
		}
	}
	for _, m := range importClauseRe.FindAllStringSubmatch(text, -1) {
		if len(m) < 3 {
			continue
		}
		clause := strings.TrimSpace(m[1])
		im := strings.TrimSpace(m[2])
		if clause == "" || im == "" {
			continue
		}
		clause = strings.TrimSuffix(clause, ",")
		parts := strings.Split(clause, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}"))
				for _, item := range strings.Split(body, ",") {
					item = strings.TrimSpace(item)
					if item == "" {
						continue
					}
					sub := namedImportItemRe.Split(item, 2)
					token := strings.TrimSpace(sub[0])
					if len(sub) == 2 {
						token = strings.TrimSpace(sub[1])
					}
					if token != "" {
						aliases[token] = im
					}
				}
				continue
			}
			if strings.HasPrefix(part, "* as ") {
				token := strings.TrimSpace(strings.TrimPrefix(part, "* as "))
				if token != "" {
					aliases[token] = im
				}
				continue
			}
			aliases[part] = im
		}
	}

	return out, aliases
}

func collectJSBlocks(root *sitter.Node, src []byte) []model.Block {
	var out []model.Block
	walkNamed(root, func(n *sitter.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case "function_declaration", "function_expression", "arrow_function", "generator_function_declaration":
			name := functionNodeName(n, src)
			if name == "" {
				name = "anonymous@" + nodeLocation(n)
			}
			b := model.Block{
				Name:       name,
				Kind:       model.BlockFunc,
				StartLine:  int(n.StartPosition().Row) + 1,
				EndLine:    int(n.EndPosition().Row) + 1,
				Complexity: jsCyclomatic(n, src),
				Calls:      collectJSCalls(n, src),
			}
			out = append(out, b)
		case "method_definition":
			name := nodeFieldText(n, "name", src)
			if name == "" {
				name = "method@" + nodeLocation(n)
			}
			b := model.Block{
				Name:       name,
				Kind:       model.BlockMethod,
				Receiver:   parentClassName(n, src),
				StartLine:  int(n.StartPosition().Row) + 1,
				EndLine:    int(n.EndPosition().Row) + 1,
				Complexity: jsCyclomatic(n, src),
				Calls:      collectJSCalls(n, src),
			}
			out = append(out, b)
		case "class_declaration":
			name := nodeFieldText(n, "name", src)
			if name == "" {
				name = "class@" + nodeLocation(n)
			}
			out = append(out, model.Block{
				Name:       name,
				Kind:       model.BlockType,
				Receiver:   "class",
				StartLine:  int(n.StartPosition().Row) + 1,
				EndLine:    int(n.EndPosition().Row) + 1,
				Complexity: 0,
			})
		case "interface_declaration":
			name := nodeFieldText(n, "name", src)
			if name == "" {
				name = "interface@" + nodeLocation(n)
			}
			out = append(out, model.Block{
				Name:       name,
				Kind:       model.BlockType,
				Receiver:   "interface",
				StartLine:  int(n.StartPosition().Row) + 1,
				EndLine:    int(n.EndPosition().Row) + 1,
				Complexity: 0,
			})
		case "type_alias_declaration":
			name := nodeFieldText(n, "name", src)
			if name == "" {
				name = "type@" + nodeLocation(n)
			}
			out = append(out, model.Block{
				Name:       name,
				Kind:       model.BlockType,
				Receiver:   "type",
				StartLine:  int(n.StartPosition().Row) + 1,
				EndLine:    int(n.EndPosition().Row) + 1,
				Complexity: 0,
			})
		}
	})
	return out
}

func collectJSLiterals(root *sitter.Node, src []byte) []model.Literal {
	var out []model.Literal
	walkNamed(root, func(n *sitter.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case "string":
			txt := strings.TrimSpace(nodeText(n, src))
			txt = strings.TrimPrefix(txt, "'")
			txt = strings.TrimPrefix(txt, "\"")
			txt = strings.TrimSuffix(txt, "'")
			txt = strings.TrimSuffix(txt, "\"")
			if txt == "" {
				return
			}
			out = append(out, model.Literal{
				Kind:  model.LitString,
				Value: txt,
				Line:  int(n.StartPosition().Row) + 1,
			})
		case "number":
			txt := strings.TrimSpace(nodeText(n, src))
			if txt == "" {
				return
			}
			out = append(out, model.Literal{
				Kind:  model.LitNumber,
				Value: txt,
				Line:  int(n.StartPosition().Row) + 1,
			})
		}
	})
	return out
}

func collectJSCalls(node *sitter.Node, src []byte) []model.Call {
	var out []model.Call
	walkNamed(node, func(n *sitter.Node) {
		if n == nil {
			return
		}
		if n.Kind() != "call_expression" && n.Kind() != "new_expression" {
			return
		}
		targetNode := n.ChildByFieldName("function")
		if targetNode == nil {
			targetNode = n.ChildByFieldName("constructor")
		}
		if targetNode == nil {
			return
		}
		target := strings.TrimSpace(nodeText(targetNode, src))
		if target == "" {
			return
		}
		target = strings.ReplaceAll(target, "\n", " ")
		target = strings.Join(strings.Fields(target), " ")
		if len(target) > 80 {
			target = target[:80]
		}
		out = append(out, model.Call{Target: target})
	})
	return out
}

func jsCyclomatic(node *sitter.Node, src []byte) int {
	score := 1
	walkNamed(node, func(n *sitter.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case "if_statement", "for_statement", "for_in_statement", "while_statement",
			"do_statement", "switch_case", "catch_clause", "conditional_expression":
			score++
		case "logical_expression":
			txt := nodeText(n, src)
			score += strings.Count(txt, "&&")
			score += strings.Count(txt, "||")
		}
	})
	return score
}

func functionNodeName(n *sitter.Node, src []byte) string {
	if n == nil {
		return ""
	}
	if name := nodeFieldText(n, "name", src); name != "" {
		return name
	}
	parent := n.Parent()
	if parent == nil {
		return ""
	}
	switch parent.Kind() {
	case "lexical_declaration", "variable_declarator":
		if name := nodeFieldText(parent, "name", src); name != "" {
			return name
		}
	}
	return ""
}

func parentClassName(n *sitter.Node, src []byte) string {
	if n == nil {
		return ""
	}
	cur := n.Parent()
	for cur != nil {
		if cur.Kind() == "class_declaration" {
			if name := nodeFieldText(cur, "name", src); name != "" {
				return name
			}
		}
		cur = cur.Parent()
	}
	return ""
}

func nodeFieldText(node *sitter.Node, field string, src []byte) string {
	if node == nil {
		return ""
	}
	n := node.ChildByFieldName(field)
	if n == nil {
		return ""
	}
	return strings.TrimSpace(nodeText(n, src))
}

func nodeLocation(n *sitter.Node) string {
	if n == nil {
		return "0"
	}
	return strconv.Itoa(int(n.StartPosition().Row) + 1)
}

func walkNamed(root *sitter.Node, fn func(*sitter.Node)) {
	if root == nil {
		return
	}
	fn(root)
	for i := uint(0); i < root.NamedChildCount(); i++ {
		ch := root.NamedChild(i)
		if ch == nil {
			continue
		}
		walkNamed(ch, fn)
	}
}

func nodeText(n *sitter.Node, src []byte) string {
	if n == nil {
		return ""
	}
	start := int(n.StartByte())
	end := int(n.EndByte())
	if start < 0 || end < 0 || start >= len(src) || end > len(src) || start >= end {
		return ""
	}
	return string(src[start:end])
}

func countLines(src []byte) int {
	if len(src) == 0 {
		return 0
	}
	return strings.Count(string(src), "\n") + 1
}
