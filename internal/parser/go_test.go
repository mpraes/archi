package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/mpraes/archi/internal/model"
)

func TestGoParserHelpers(t *testing.T) {
	g := &goParser{root: t.TempDir(), modulePath: "example.com/mod"}
	path := filepath.Join(g.root, "pkg", "file.go")
	if mod := g.moduleFor(path); mod != "example.com/mod/pkg" {
		t.Fatalf("moduleFor = %q", mod)
	}
	if !g.isInternal("example.com/mod/pkg") || g.isInternal("fmt") {
		t.Fatal("isInternal mismatch")
	}
	if lastElem("example.com/mod/pkg") != "pkg" {
		t.Fatal("lastElem")
	}
}

func TestReceiverAndSelectorText(t *testing.T) {
	src := `package p
type S struct{}
func (s *S) M() { x.Foo() }
func F() { if a && b { for range c {} } }
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var recv ast.Expr
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv != nil {
			recv = fn.Recv.List[0].Type
			break
		}
	}
	if receiverTypeName(recv) != "S" {
		t.Fatalf("receiver = %q", receiverTypeName(recv))
	}
}

func TestCycloAndCallVisitor(t *testing.T) {
	src := `package p
func F(x int) {
	if x > 0 && x < 10 {
		for range []int{1} {
			switch x {
			case 1:
				foo.Bar()
			}
		}
	}
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		cv := &cycloVisitor{}
		ast.Walk(cv, fn.Body)
		if cv.count < 3 {
			t.Fatalf("cyclo = %d", cv.count)
		}
		callV := &callVisitor{modulePath: "example.com/mod", modName: "example.com/mod/p"}
		ast.Walk(callV, fn.Body)
		if len(callV.calls) == 0 {
			t.Fatal("expected calls")
		}
	}
}

func TestResolveCallTarget(t *testing.T) {
	g := &goParser{modulePath: "example.com/mod"}
	byName := map[string]*model.Module{
		"example.com/mod/alpha": {Name: "example.com/mod/alpha"},
		"example.com/mod/p":     {Name: "example.com/mod/p"},
	}
	if got := g.resolveCallTarget("alpha.Run", byName, "example.com/mod/p"); got != "example.com/mod/alpha" {
		t.Fatalf("resolve = %q", got)
	}
}
