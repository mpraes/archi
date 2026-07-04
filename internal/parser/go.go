package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mpraes/archi/internal/model"
)

type goParser struct {
	root       string
	modulePath string
	warnings   *[]string

	files []model.File
	modIdx map[string]int // module name -> index in modules slice
	modulesList []model.Module
}

func (g *goParser) parseFile(path string) (model.File, error) {
	fset := token.NewFileSet()
	// ParseFile with AllErrors but recover; we collect errors and continue.
	var errs []string
	astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		// On hard syntax failure, record warning and skip file (RNF-003).
		if astFile == nil {
			return model.File{}, err
		}
		errs = append(errs, err.Error())
	}

	f := model.File{Path: path}
	if astFile != nil {
		f.Module = g.moduleFor(path)
		tf := fset.File(astFile.Pos())
		f.Lines = tf.LineCount()
		f.Imports = g.imports(astFile)
		f.Blocks, f.Literals, f.Errors = g.scan(fset, astFile, path)
		f.Errors = append(f.Errors, errs...)
	}
	g.files = append(g.files, f)
	return f, nil
}

func (g *goParser) imports(f *ast.File) []string {
	out := make([]string, 0, len(f.Imports))
	for _, im := range f.Imports {
		v, err := strconv.Unquote(im.Path.Value)
		if err != nil {
			v = strings.Trim(im.Path.Value, "\"")
		}
		out = append(out, v)
	}
	return out
}

func (g *goParser) scan(fset *token.FileSet, f *ast.File, path string) ([]model.Block, []model.Literal, []string) {
	var blocks []model.Block
	var lits []model.Literal
	var errs []string

	// Top-level declarations.
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			b := model.Block{
				Name:       d.Name.Name,
				StartLine:  fset.Position(d.Pos()).Line,
				EndLine:    fset.Position(d.End()).Line,
			}
			if d.Recv != nil && len(d.Recv.List) > 0 {
				b.Kind = model.BlockMethod
				b.Receiver = receiverTypeName(d.Recv.List[0].Type)
			} else {
				b.Kind = model.BlockFunc
			}
			if d.Body != nil {
				v := &cycloVisitor{}
				ast.Walk(v, d.Body)
				b.Complexity = v.count + 1
				cv := &callVisitor{modulePath: g.modulePath, modName: g.moduleFor(path)}
				ast.Walk(cv, d.Body)
				b.Calls = cv.calls
			}
			blocks = append(blocks, b)
		case *ast.GenDecl:
			// Count type declarations: interfaces (abstracts) vs others (concretes).
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				b := model.Block{
					Name:       ts.Name.Name,
					Kind:       model.BlockType,
					StartLine:  fset.Position(ts.Pos()).Line,
					EndLine:    fset.Position(ts.End()).Line,
					Complexity: 0,
				}
				switch ts.Type.(type) {
				case *ast.InterfaceType:
					b.Receiver = "interface"
				case *ast.StructType:
					b.Receiver = "struct"
				}
				blocks = append(blocks, b)
			}
		}
	}

	// Literals (for CoM detection).
	ast.Inspect(f, func(n ast.Node) bool {
		switch li := n.(type) {
		case *ast.BasicLit:
			l := model.Literal{Line: fset.Position(li.Pos()).Line}
			switch li.Kind {
			case token.STRING:
				l.Kind = model.LitString
				if v, err := strconv.Unquote(li.Value); err == nil {
					l.Value = v
				} else {
					l.Value = li.Value
				}
			case token.INT, token.FLOAT:
				l.Kind = model.LitNumber
				l.Value = li.Value
			default:
				return true
			}
			lits = append(lits, l)
		}
		return true
	})

	return blocks, lits, errs
}

func receiverTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	}
	return ""
}

// callVisitor collects function/method call target textual identifiers.
type callVisitor struct {
	modulePath string
	modName    string
	calls      []model.Call
}

func (c *callVisitor) Visit(n ast.Node) ast.Visitor {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return c
	}
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		c.calls = append(c.calls, model.Call{Target: fun.Name})
	case *ast.SelectorExpr:
		// pkg.Foo or x.Method — record the textual selector chain.
		t := selectorText(fun)
		c.calls = append(c.calls, model.Call{Target: t})
	}
	return c
}

func selectorText(s *ast.SelectorExpr) string {
	var b strings.Builder
	var write func(ast.Expr)
	write = func(e ast.Expr) {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			write(x.X)
			b.WriteString(".")
			b.WriteString(x.Sel.Name)
		case *ast.Ident:
			b.WriteString(x.Name)
		case *ast.CallExpr:
			write(x.Fun)
		default:
			b.WriteString("?")
		}
	}
	write(s.X)
	b.WriteString(".")
	b.WriteString(s.Sel.Name)
	return b.String()
}

// cycloVisitor counts decision points for cyclomatic complexity.
type cycloVisitor struct {
	count int
}

func (v *cycloVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt,
		*ast.TypeSwitchStmt, *ast.CaseClause, *ast.LabeledStmt:
		v.count++
	case *ast.BinaryExpr:
		// && / || add one path each.
		if n.(*ast.BinaryExpr).Op == token.LAND || n.(*ast.BinaryExpr).Op == token.LOR {
			v.count++
		}
	}
	return v
}

// moduleFor maps a file path to its package module name (module path + relative dir).
func (g *goParser) moduleFor(path string) string {
	rel, err := filepath.Rel(g.root, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	dir := filepath.Dir(rel)
	if dir == "." {
		if g.modulePath != "" {
			return g.modulePath
		}
		return "root"
	}
	if g.modulePath != "" {
		return g.modulePath + "/" + filepath.ToSlash(dir)
	}
	return filepath.ToSlash(dir)
}

func (g *goParser) modules() []model.Module {
	g.modIdx = map[string]int{}
	for _, f := range g.files {
		idx, ok := g.modIdx[f.Module]
		if !ok {
			g.modulesList = append(g.modulesList, model.Module{
				Name:    f.Module,
				Path:    filepath.Dir(f.Path),
				Imports: map[string]struct{}{},
			})
			idx = len(g.modulesList) - 1
			g.modIdx[f.Module] = idx
		}
		m := &g.modulesList[idx]
		m.Files++
		for _, im := range f.Imports {
			if g.isInternal(im) {
				m.Imports[im] = struct{}{}
			}
		}
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
	return g.modulesList
}

var blockedStdPackages = []string{"math/rand"}

// isInternal reports whether an import path is a project package.
func (g *goParser) isInternal(im string) bool {
	if g.modulePath == "" {
		return false
	}
	if !strings.HasPrefix(im, g.modulePath) {
		return false
	}
	rest := strings.TrimPrefix(im, g.modulePath)
	if rest == "" {
		return true
	}
	if rest[0] != '/' {
		return false
	}
	// Exclude blocked stdlib
	for _, b := range blockedStdPackages {
		if im == b {
			return false
		}
	}
	return true
}

func (g *goParser) resolve(p *model.Program) {
	// Build index of modules by name.
	byName := map[string]*model.Module{}
	for i := range p.Modules {
		byName[p.Modules[i].Name] = &p.Modules[i]
	}
	for i := range p.Files {
		f := &p.Files[i]
		for bi := range f.Blocks {
			b := &f.Blocks[bi]
			for ci := range b.Calls {
				c := &b.Calls[ci]
				if mod := g.resolveCallTarget(c.Target, byName, f.Module); mod != "" {
					c.ResolvedModule = mod
				}
			}
		}
	}
}

// resolveCallTarget maps a textual call target to an internal module.
//
// Heuristic: "pkg.Foo" → if "pkg" prefix matches another module's name suffix → resolved.
// This is best-effort: callers passing alias names are not resolved; unresolved
// calls are simply not counted (they are external/stdlib).
func (g *goParser) resolveCallTarget(target string, byName map[string]*model.Module, from string) string {
	if target == "" {
		return ""
	}
	// "X.Y" — try the directory-base name of X against module name suffixes.
	if dot := strings.Index(target, "."); dot > 0 {
		pkgPart := target[:dot]
		// exact short match: pkgPart equals last path element of a module name.
		for _, m := range byName {
			if m.Name == from {
				continue
			}
			if lastElem(m.Name) == pkgPart {
				return m.Name
			}
		}
	}
	return ""
}

func lastElem(importPath string) string {
	i := strings.LastIndex(importPath, "/")
	if i < 0 {
		return importPath
	}
	return importPath[i+1:]
}