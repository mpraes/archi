package parser

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"

	"github.com/mpraes/archi/internal/model"
)

var (
	pyImportRe         = regexp.MustCompile(`(?m)^\s*import\s+([a-zA-Z0-9_.,\s]+)$`)
	pyFromImportRe     = regexp.MustCompile(`(?m)^\s*from\s+([\.a-zA-Z0-9_]+)\s+import\s+([a-zA-Z0-9_.*, \t]+)$`)
	pyAsSplitRe        = regexp.MustCompile(`\s+as\s+`)
	pyDotCollapseSplit = regexp.MustCompile(`\.+`)
)

type pyParser struct {
	root     string
	warnings *[]string

	files         []model.File
	modulesList   []model.Module
	aliasesByFile map[string]map[string]string
}

func (p *pyParser) parseFile(filePath string) (model.File, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return model.File{}, err
	}

	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(sitter.NewLanguage(tree_sitter_python.Language())); err != nil {
		return model.File{}, err
	}

	tree := parser.Parse(src, nil)
	defer tree.Close()
	root := tree.RootNode()

	imports, aliases := parsePythonImportsDetails(src)
	if p.aliasesByFile == nil {
		p.aliasesByFile = map[string]map[string]string{}
	}
	p.aliasesByFile[filePath] = aliases

	f := model.File{
		Path:     filePath,
		Module:   p.moduleFor(filePath),
		Lines:    countLines(src),
		Imports:  imports,
		Blocks:   collectPythonBlocks(root, src),
		Literals: collectPythonLiterals(root, src),
	}
	if root != nil && root.HasError() {
		f.Errors = append(f.Errors, "syntax errors detected; partial tree-sitter parse used")
	}
	p.files = append(p.files, f)
	return f, nil
}

func (p *pyParser) modules() []model.Module {
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
				Language: "py",
				Imports:  map[string]struct{}{},
			})
			idx = len(p.modulesList) - 1
			indexByName[f.Module] = idx
		}
		m := &p.modulesList[idx]
		m.Files++
		rawImportsByModule[f.Module] = append(rawImportsByModule[f.Module], f.Imports...)
		for _, b := range f.Blocks {
			if b.Kind == model.BlockType {
				m.Concretes++
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
			target := p.normalizePyImport(im, m.Name)
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

func (p *pyParser) resolve(prog *model.Program) {
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
				targetToken := callRootToken(b.Calls[ci].Target)
				if rawTarget, ok := aliases[targetToken]; ok {
					norm := p.normalizePyImport(rawTarget, f.Module)
					if resolved := normalizeKnownModule(norm, byName); resolved != "" {
						b.Calls[ci].ResolvedModule = resolved
						continue
					}
				}
				b.Calls[ci].ResolvedModule = resolveCallByToken(b.Calls[ci].Target, f.Module, byName)
			}
		}
	}
}

func resolveCallByToken(target, from string, modules map[string]struct{}) string {
	token := callRootToken(target)
	if token == "" || token == "self" || token == "cls" {
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

func normalizeKnownModule(candidate string, modules map[string]struct{}) string {
	candidate = strings.Trim(candidate, "/")
	if candidate == "" {
		return ""
	}
	if _, ok := modules[candidate]; ok {
		return candidate
	}
	// from package.sub.symbol -> keep collapsing suffix until a known module appears.
	cur := candidate
	for {
		if i := strings.LastIndex(cur, "/"); i > 0 {
			cur = cur[:i]
			if _, ok := modules[cur]; ok {
				return cur
			}
			continue
		}
		break
	}
	// from root-prefixed imports (e.g. "src/utils") -> try dropping leading segments.
	cur = candidate
	for {
		if i := strings.Index(cur, "/"); i > 0 {
			cur = cur[i+1:]
			if _, ok := modules[cur]; ok {
				return cur
			}
			continue
		}
		break
	}
	return ""
}

func (p *pyParser) moduleFor(filePath string) string {
	rel, err := filepath.Rel(p.root, filePath)
	if err != nil {
		return "root"
	}
	rel = filepath.ToSlash(rel)
	if strings.HasSuffix(rel, "/__init__.py") {
		dir := path.Dir(rel)
		if dir == "." || dir == "" {
			return "root"
		}
		return dir
	}
	dir := path.Dir(rel)
	if dir == "." || dir == "" {
		return "root"
	}
	return dir
}

func (p *pyParser) normalizePyImport(raw, fromModule string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, ".") {
		return resolveRelativePyImport(raw, fromModule)
	}
	return strings.ReplaceAll(raw, ".", "/")
}

func resolveRelativePyImport(raw, fromModule string) string {
	dots := 0
	for dots < len(raw) && raw[dots] == '.' {
		dots++
	}
	tail := strings.TrimPrefix(raw, strings.Repeat(".", dots))
	base := fromModule
	for i := 1; i < dots; i++ {
		if base == "root" || base == "" || base == "." {
			base = "root"
			break
		}
		base = path.Dir(base)
		if base == "." {
			base = "root"
		}
	}
	tail = strings.ReplaceAll(tail, ".", "/")
	switch {
	case tail == "" && base != "":
		return base
	case base == "" || base == "." || base == "root":
		if tail == "" {
			return "root"
		}
		return tail
	default:
		return strings.TrimPrefix(path.Clean(base+"/"+tail), "./")
	}
}

func parsePythonImportsDetails(src []byte) ([]string, map[string]string) {
	text := string(src)
	seen := map[string]struct{}{}
	var out []string
	aliases := map[string]string{}

	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	for _, m := range pyImportRe.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		for _, item := range strings.Split(m[1], ",") {
			mod, token := parsePyImportItem(item)
			if mod == "" {
				continue
			}
			add(mod)
			if token != "" {
				aliases[token] = mod
			}
		}
	}
	for _, m := range pyFromImportRe.FindAllStringSubmatch(text, -1) {
		if len(m) < 3 {
			continue
		}
		fromMod := strings.TrimSpace(m[1])
		if fromMod == "" {
			continue
		}
		add(fromMod)
		symbols := strings.Split(m[2], ",")
		for _, item := range symbols {
			symbol, token := parsePyImportItem(item)
			if symbol == "" || symbol == "*" {
				continue
			}
			if token == "" {
				token = symbol
			}
			aliases[token] = fromMod + "." + symbol
		}
	}
	return out, aliases
}

func parsePyImportItem(item string) (module string, token string) {
	item = strings.TrimSpace(item)
	if item == "" {
		return "", ""
	}
	parts := pyAsSplitRe.Split(item, 2)
	module = strings.TrimSpace(parts[0])
	if module == "" {
		return "", ""
	}
	if len(parts) == 2 {
		token = strings.TrimSpace(parts[1])
		return module, token
	}
	// Default token for "import pkg.sub" is "pkg".
	head := pyDotCollapseSplit.Split(module, 2)
	if len(head) > 0 {
		token = strings.TrimSpace(head[0])
	}
	if token == "" {
		token = module
	}
	return module, token
}

func collectPythonBlocks(root *sitter.Node, src []byte) []model.Block {
	var out []model.Block
	walkNamed(root, func(n *sitter.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case "function_definition", "async_function_definition":
			name := nodeFieldText(n, "name", src)
			if name == "" {
				name = "function@" + nodeLocation(n)
			}
			kind := model.BlockFunc
			recv := ""
			if className := parentPythonClassName(n, src); className != "" {
				kind = model.BlockMethod
				recv = className
			}
			out = append(out, model.Block{
				Name:       name,
				Kind:       kind,
				Receiver:   recv,
				StartLine:  int(n.StartPosition().Row) + 1,
				EndLine:    int(n.EndPosition().Row) + 1,
				Complexity: pythonCyclomatic(n, src),
				Calls:      collectPythonCalls(n, src),
			})
		case "class_definition":
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
		}
	})
	return out
}

func collectPythonLiterals(root *sitter.Node, src []byte) []model.Literal {
	var out []model.Literal
	walkNamed(root, func(n *sitter.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case "string":
			txt := strings.TrimSpace(nodeText(n, src))
			txt = strings.Trim(txt, "\"'")
			if txt == "" {
				return
			}
			out = append(out, model.Literal{Kind: model.LitString, Value: txt, Line: int(n.StartPosition().Row) + 1})
		case "integer", "float":
			txt := strings.TrimSpace(nodeText(n, src))
			if txt == "" {
				return
			}
			out = append(out, model.Literal{Kind: model.LitNumber, Value: txt, Line: int(n.StartPosition().Row) + 1})
		}
	})
	return out
}

func collectPythonCalls(node *sitter.Node, src []byte) []model.Call {
	var out []model.Call
	walkNamed(node, func(n *sitter.Node) {
		if n == nil || n.Kind() != "call" {
			return
		}
		fn := n.ChildByFieldName("function")
		if fn == nil {
			return
		}
		target := strings.TrimSpace(nodeText(fn, src))
		if target == "" {
			return
		}
		target = strings.Join(strings.Fields(strings.ReplaceAll(target, "\n", " ")), " ")
		out = append(out, model.Call{Target: target})
	})
	return out
}

func pythonCyclomatic(node *sitter.Node, src []byte) int {
	score := 1
	walkNamed(node, func(n *sitter.Node) {
		if n == nil {
			return
		}
		switch n.Kind() {
		case "if_statement", "for_statement", "while_statement", "except_clause",
			"conditional_expression", "match_case":
			score++
		case "boolean_operator":
			txt := nodeText(n, src)
			score += strings.Count(txt, " and ")
			score += strings.Count(txt, " or ")
		}
	})
	return score
}

func parentPythonClassName(n *sitter.Node, src []byte) string {
	for cur := n.Parent(); cur != nil; cur = cur.Parent() {
		if cur.Kind() == "class_definition" {
			return nodeFieldText(cur, "name", src)
		}
	}
	return ""
}
